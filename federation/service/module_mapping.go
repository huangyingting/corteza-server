package service

import (
	"context"

	composeService "github.com/cortezaproject/corteza-server/compose/service"
	"github.com/cortezaproject/corteza-server/federation/types"
	"github.com/cortezaproject/corteza-server/pkg/actionlog"
	"github.com/cortezaproject/corteza-server/store"
)

type (
	moduleMapping struct {
		store     store.Storer
		module    composeService.ModuleService
		namespace composeService.NamespaceService
		actionlog actionlog.Recorder
	}

	ModuleMappingService interface {
		Find(ctx context.Context, filter types.ModuleMappingFilter) (types.ModuleMappingSet, types.ModuleMappingFilter, error)
		FindByID(ctx context.Context, federationModuleID uint64) (*types.ModuleMapping, error)
		Create(ctx context.Context, new *types.ModuleMapping) (*types.ModuleMapping, error)
		Update(ctx context.Context, updated *types.ModuleMapping) (*types.ModuleMapping, error)
		// FindByAny(ctx context.Context, nodeID uint64, identifier interface{}) (*types.ExposedModule, error)
		// DeleteByID(ctx context.Context, federationModuleID uint64) error
	}

	moduleMappingUpdateHandler func(ctx context.Context, c *types.ModuleMapping) (bool, bool, error)
)

func ModuleMapping() ModuleMappingService {
	return &moduleMapping{
		store:     DefaultStore,
		actionlog: DefaultActionlog,
		module:    composeService.DefaultModule,
		namespace: composeService.DefaultNamespace,
	}
}

// func (svc moduleMapping) updater(ctx context.Context, federationModuleID, composeModuleID uint64, action func(...*moduleMappingActionProps) *moduleMappingAction, fn moduleUpdateHandler) (*types.ModuleMapping, error) {
// 	var (
// 		m      *types.ModuleMapping
// 		aProps = &moduleMappingActionProps{mapping: &types.ModuleMapping{FederationModuleID: federationModuleID, ComposeModuleID: composeModuleID}}
// 		err    error
// 	)

// 	err = store.Tx(ctx, svc.store, func(ctx context.Context, s store.Storer) (err error) {
// 		if m, err = svc.store.LookupFederationModuleMappingByFederationModuleIDComposeModuleID(ctx, federationModuleID, composeModuleID); err != nil {
// 			return err
// 		}

// 		// TODO - handle node id also
// 		if moduleChanged, fieldsChanged, err = fn(ctx, n, m); err != nil {
// 			return err
// 		}

// 		return err
// 	})

// 	return m, svc.recordAction(ctx, aProps, action, err)
// }

// func (svc moduleMapping) DeleteByID(ctx context.Context, federationModuleID, composeModuleID uint64) error {
// 	_, err := svc.updater(ctx, nodeID, moduleID, ModuleMappingActionDelete(), svc.handleDelete)

// 	return err
// }

// func (svc moduleMapping) handleDelete(ctx context.Context, m *types.ModuleMapping) (bool, bool, error) {
// 	if err := store.DeleteFederationModuleMapping(ctx, svc.store, m); err != nil {
// 		return false, false, err
// 	}

// 	return false, false, nil
// }

func (svc moduleMapping) FindByID(ctx context.Context, federationModuleID uint64) (m *types.ModuleMapping, err error) {
	err = func() error {
		if m, err = store.LookupFederationModuleMappingByFederationModuleID(ctx, svc.store, federationModuleID); err != nil {
			return err
		}

		return nil
	}()

	return m, err
}

func (svc moduleMapping) Find(ctx context.Context, filter types.ModuleMappingFilter) (set types.ModuleMappingSet, f types.ModuleMappingFilter, err error) {
	err = func() error {
		if set, f, err = store.SearchFederationModuleMappings(ctx, svc.store, filter); err != nil {
			return err
		}

		return nil
	}()

	return set, f, err
}

func (svc moduleMapping) Create(ctx context.Context, new *types.ModuleMapping) (*types.ModuleMapping, error) {
	var (
		aProps = &moduleMappingActionProps{created: new}
	)

	err := store.Tx(ctx, svc.store, func(ctx context.Context, s store.Storer) (err error) {
		// TODO
		// if !svc.ac.CanCreateFederationExposedModule(ctx, ns) {
		// 	return ExposedModuleErrNotAllowedToCreate()
		// }

		if _, err := svc.namespace.With(ctx).FindByID(new.ComposeNamespaceID); err != nil {
			return ModuleMappingErrComposeNamespaceNotFound()
		}

		if _, err := svc.module.With(ctx).FindByID(new.ComposeNamespaceID, new.ComposeModuleID); err != nil {
			return ModuleMappingErrComposeModuleNotFound()
		}

		// Check for federation module - compose.Module combo
		if err = svc.uniqueCheck(ctx, new); err != nil {
			return err
		}

		if err = store.CreateFederationModuleMapping(ctx, s, new); err != nil {
			return err
		}

		return nil
	})

	return new, svc.recordAction(ctx, aProps, ModuleMappingActionCreate, err)
}

func (svc moduleMapping) Update(ctx context.Context, updated *types.ModuleMapping) (*types.ModuleMapping, error) {
	var (
		aProps = &moduleMappingActionProps{changed: updated}
	)

	err := store.Tx(ctx, svc.store, func(ctx context.Context, s store.Storer) (err error) {
		// TODO
		// if !svc.ac.CanCreateFederationExposedModule(ctx, ns) {
		// 	return ExposedModuleErrNotAllowedToCreate()
		// }

		if _, err := svc.namespace.With(ctx).FindByID(updated.ComposeNamespaceID); err != nil {
			return ModuleMappingErrComposeNamespaceNotFound()
		}

		if _, err := svc.module.With(ctx).FindByID(updated.ComposeNamespaceID, updated.ComposeModuleID); err != nil {
			return ModuleMappingErrComposeModuleNotFound()
		}

		if err = store.UpdateFederationModuleMapping(ctx, s, updated); err != nil {
			return err
		}

		return nil
	})

	return updated, svc.recordAction(ctx, aProps, ModuleMappingActionUpdate, err)
}

func (svc moduleMapping) uniqueCheck(ctx context.Context, m *types.ModuleMapping) (err error) {
	f := types.ModuleMappingFilter{
		FederationModuleID: m.FederationModuleID,
		ComposeModuleID:    m.ComposeModuleID,
		ComposeNamespaceID: m.ComposeNamespaceID,
	}

	if set, _, err := store.SearchFederationModuleMappings(ctx, svc.store, f); len(set) > 0 && err == nil {
		return ModuleMappingErrModuleMappingExists()
	} else if err != nil {
		return err
	}

	return err
}
