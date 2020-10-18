package action

//import (
//	"github.com/cortezaproject/corteza-server/molding"
//)
//
//type (
//	sequenceNextValConfig struct {
//		Name      string
//		Start     int64
//		Increment int64
//	}
//
//	sequenceNextVal struct {
//		// sequence reference
//		// @todo implementation
//		name string
//
//		// in case we need to define it, where do we start?
//		// default: 0
//		start int64
//
//		// in case we need to define it, what is the interval?
//		// this needs to be <>0 to trigger sequence definition
//		increment int64
//
//		// temp in-memory store for sequence
//		current int64
//	}
//)
//
//var (
//	_ molding.Executor = &sequenceNextVal{}
//)
//
//func NewSequenceNextVal(cfg sequenceNextValConfig) (*sequenceNextVal, error) {
//	var (
//		seq = &sequenceNextVal{}
//	)
//
//	seq.name = cfg.Name
//	seq.start = cfg.Start
//	seq.increment = cfg.Increment
//
//	// @todo get the actual "current" value from somewhere
//	seq.current = cfg.Start
//
//	return seq, nil
//}
//
//func (a *sequenceNextVal) Execute(_ molding.Variables) (molding.Variables, error) {
//	a.current += a.increment
//	return molding.Variables{
//		"value": a.current,
//	}, nil
//}
//
//func (a *sequenceNextVal) Assign() map[string]string {
//	panic("implement me")
//}
//
//func (a *sequenceNextVal) Expects() molding.Params { return nil }
