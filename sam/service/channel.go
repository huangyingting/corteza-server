package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/titpetric/factory"

	"github.com/crusttech/crust/sam/repository"
	"github.com/crusttech/crust/sam/types"
)

type (
	channel struct {
		db  *factory.DB
		ctx context.Context

		channel repository.ChannelRepository
		message repository.MessageRepository
	}

	ChannelService interface {
		With(ctx context.Context) ChannelService

		FindByID(channelID uint64) (*types.Channel, error)
		Find(filter *types.ChannelFilter) ([]*types.Channel, error)
		FindByMembership() (rval []*types.Channel, err error)

		Create(channel *types.Channel) (*types.Channel, error)
		Update(channel *types.Channel) (*types.Channel, error)

		deleter
		archiver
	}

	//channelSecurity interface {
	//	CanRead(ch *types.Channel) bool
	//}
)

func Channel() *channel {
	svc := (&channel{}).With(context.Background()).(*channel)
	//svc.sec.ch = ChannelSecurity(svc.channel)
	return svc
}

func (svc *channel) With(ctx context.Context) ChannelService {
	db := repository.DB(ctx)
	return &channel{
		db:      db,
		ctx:     ctx,
		channel: repository.Channel(ctx, db),
		message: repository.Message(ctx, db),
	}
}

func (svc *channel) FindByID(id uint64) (ch *types.Channel, err error) {
	ch, err = svc.channel.FindChannelByID(id)
	if err != nil {
		return
	}

	//if !svc.sec.ch.CanRead(ch) {
	//	return nil, errors.New("Not allowed to access channel")
	//}

	return
}

func (svc *channel) Find(filter *types.ChannelFilter) ([]*types.Channel, error) {
	// @todo: permission check to return only channels that channel has access to
	if cc, err := svc.channel.FindChannels(filter); err != nil {
		return nil, err
	} else {
		return cc, svc.preloadMembers(cc)
	}
}

func (svc *channel) preloadMembers(set types.ChannelSet) error {
	// @todo implement
	return nil
}

// Returns all channels with membership info
func (svc *channel) FindByMembership() (rval []*types.Channel, err error) {
	return rval, svc.db.Transaction(func() error {
		var chMemberId = repository.Identity(svc.ctx)
		var mm []*types.ChannelMember

		if mm, err = svc.channel.FindChannelsMembershipsByMemberId(chMemberId); err != nil {
			return err
		}

		if rval, err = svc.channel.FindChannels(nil); err != nil {
			return err
		}

		for _, m := range mm {
			for _, c := range rval {
				if c.ID == m.ChannelID {
					c.Member = m
				}
			}
		}

		return nil
	})
}

func (svc *channel) Create(in *types.Channel) (out *types.Channel, err error) {
	// @todo: [SECURITY] permission check if user can add channel

	return out, svc.db.Transaction(func() (err error) {
		var msg *types.Message

		// @todo get organisation from somewhere
		var organisationID uint64 = 0

		var chCreatorID = repository.Identity(svc.ctx)

		// @todo [SECURITY] check if channel topic can be set
		if in.Topic != "" && false {
			return errors.New("Not allowed to set channel topic")
		}

		// @todo [SECURITY] check if user can create public channels
		if in.Type == types.ChannelTypePublic && false {
			return errors.New("Not allowed to create public channels")
		}

		// @todo [SECURITY] check if user can create private channels
		if in.Type == types.ChannelTypePrivate && false {
			return errors.New("Not allowed to create public channels")
		}

		// @todo [SECURITY] check if user can create private channels
		if in.Type == types.ChannelTypeGroup && false {
			return errors.New("Not allowed to create group channels")
		}

		// This is a fresh channel, just copy values
		out = &types.Channel{
			Name:           in.Name,
			Topic:          in.Topic,
			Type:           in.Type,
			OrganisationID: organisationID,
			CreatorID:      chCreatorID,
		}

		// Save the channel
		if out, err = svc.channel.CreateChannel(out); err != nil {
			return
		}

		// Join current user as an member & owner
		_, err = svc.channel.AddChannelMember(&types.ChannelMember{
			ChannelID: out.ID,
			UserID:    chCreatorID,
			Type:      types.ChannelMembershipTypeOwner,
		})

		if err != nil {
			// Could not add member
			return
		}

		// Create the first message, doing this directly with repository to circumvent
		// message service constraints
		msg, err = svc.message.CreateMessage(svc.makeSystemMessage(
			out,
			"@%d created new %s channel, topic is: %s",
			chCreatorID,
			"<PRIVATE-OR-PUBLIC>",
			"<TOPIC>"))

		if err != nil {
			// Message creation failed
			return
		}

		// @todo send channel creation to the event-loop
		// @todo send msg to the event-loop
		_ = msg

		return nil
	})
}

func (svc *channel) Update(in *types.Channel) (out *types.Channel, err error) {
	return out, svc.db.Transaction(func() (err error) {
		var msgs types.MessageSet

		// @todo [SECURITY] can user access this channel?
		if out, err = svc.channel.FindChannelByID(in.ID); err != nil {
			return
		}

		if out.ArchivedAt != nil {
			return errors.New("Not allowed to edit archived channels")
		} else if out.DeletedAt != nil {
			return errors.New("Not allowed to edit deleted channels")
		}

		var chUpdatorId = repository.Identity(svc.ctx)

		// Copy values
		if out.Name != in.Name {
			// @todo [SECURITY] can we change channel's name?
			if false {
				return errors.New("Not allowed to rename channel")
			} else {
				msgs = append(msgs, svc.makeSystemMessage(
					out, "@%d renamed channel %s (was: %s)", chUpdatorId, out.Name, in.Name))
			}
			out.Name = in.Name
		}

		if out.Topic != in.Topic && true {
			// @todo [SECURITY] can we change channel's topic?
			if false {
				return errors.New("Not allowed to change channel topic")
			} else {
				msgs = append(msgs, svc.makeSystemMessage(
					out, "@%d changed channel topic: %s (was: %s)", chUpdatorId, out.Topic, in.Topic))
			}

			out.Topic = in.Topic
		}

		if out.Type != in.Type {
			// @todo [SECURITY] check if user can create public channels
			if in.Type == types.ChannelTypePublic && false {
				return errors.New("Not allowed to change type of this channel to public")
			}

			// @todo [SECURITY] check if user can create private channels
			if in.Type == types.ChannelTypePrivate && false {
				return errors.New("Not allowed to change type of this channel to private")
			}

			// @todo [SECURITY] check if user can create group channels
			if in.Type == types.ChannelTypeGroup && false {
				return errors.New("Not allowed to change type of this channel to group")
			}
		}

		// Save the updated channel
		if out, err = svc.channel.UpdateChannel(in); err != nil {
			return
		}

		// @todo send channel creation to the event-loop

		// Create the first message, doing this directly with repository to circumvent
		// message service constraints
		for _, msg := range msgs {
			if msg, err = svc.message.CreateMessage(msg); err != nil {
				// @todo send new msg to the event-loop
				return err
			}
		}

		if err != nil {
			// Message creation failed
			return
		}

		return nil
	})
}

func (svc *channel) Delete(id uint64) error {
	return svc.db.Transaction(func() (err error) {
		var userID = repository.Identity(svc.ctx)
		var ch *types.Channel

		// @todo [SECURITY] can user access this channel?
		if ch, err = svc.channel.FindChannelByID(id); err != nil {
			return
		}

		// @todo [SECURITY] can user delete this channel?

		if ch.DeletedAt != nil {
			return errors.New("Channel already deleted")
		}

		_, err = svc.message.CreateMessage(svc.makeSystemMessage(ch, "@%d deleted this channel", userID))

		return svc.channel.DeleteChannelByID(id)
	})
}

func (svc *channel) Recover(id uint64) error {
	return svc.db.Transaction(func() (err error) {
		var userID = repository.Identity(svc.ctx)
		var ch *types.Channel

		// @todo [SECURITY] can user access this channel?
		if ch, err = svc.channel.FindChannelByID(id); err != nil {
			return
		}

		// @todo [SECURITY] can user recover this channel?

		if ch.DeletedAt == nil {
			return errors.New("Channel not deleted")
		}

		_, err = svc.message.CreateMessage(svc.makeSystemMessage(ch, "@%d recovered this channel", userID))

		return svc.channel.DeleteChannelByID(id)
	})
}

func (svc *channel) Archive(id uint64) error {
	return svc.db.Transaction(func() (err error) {
		var userID = repository.Identity(svc.ctx)
		var ch *types.Channel

		// @todo [SECURITY] can user access this channel?
		if ch, err = svc.channel.FindChannelByID(id); err != nil {
			return
		}

		// @todo [SECURITY] can user archive this channel?

		if ch.ArchivedAt != nil {
			return errors.New("Channel already archived")
		}

		_, err = svc.message.CreateMessage(svc.makeSystemMessage(ch, "@%d archived this channel", userID))

		return svc.channel.ArchiveChannelByID(id)
	})
}

func (svc *channel) Unarchive(id uint64) error {
	return svc.db.Transaction(func() (err error) {
		var userID = repository.Identity(svc.ctx)
		var ch *types.Channel

		// @todo [SECURITY] can user access this channel?
		if ch, err = svc.channel.FindChannelByID(id); err != nil {
			return
		}

		// @todo [SECURITY] can user unarchive this channel?

		if ch.ArchivedAt == nil {
			return errors.New("Channel not archived")
		}

		_, err = svc.message.CreateMessage(svc.makeSystemMessage(ch, "@%d unarchived this channel", userID))

		return svc.channel.ArchiveChannelByID(id)
	})

}

func (svc *channel) makeSystemMessage(ch *types.Channel, format string, a ...interface{}) *types.Message {
	return &types.Message{
		ChannelID: ch.ID,
		Message:   fmt.Sprintf(format, a...),
	}
}

//// @todo temp location, move this somewhere else
//type (
//	nativeChannelSec struct {
//		rpo struct {
//			ch nativeChannelSecChRepo
//		}
//	}
//
//	nativeChannelSecChRepo interface {
//		FindMember(channelId uint64, userId uint64) (*types.User, error)
//	}
//)
//
//func ChannelSecurity(chRpo nativeChannelSecChRepo) channelSecurity {
//	var sec = &nativeChannelSec{}
//
//	sec.rpo.ch = chRpo
//
//	return sec
//}
//
//// Current user can read the channel if he is a member
//func (sec nativeChannelSec) CanRead(ch *types.Channel) bool {
//	// @todo check if channel is public?
//
//	var currentUserID = repository.Identity(svc.ctx)
//
//	user, err := sec.rpo.FindMember(ch.ID, currentUserID)
//
//	return err != nil && user.Valid()
//}

var _ ChannelService = &channel{}
