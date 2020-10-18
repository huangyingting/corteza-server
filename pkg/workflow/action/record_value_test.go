package action

//import (
//	"github.com/cortezaproject/corteza-server/compose/types"
//	"github.com/cortezaproject/corteza-server/molding"
//	"github.com/cortezaproject/corteza-server/pkg/id"
//	sysTypes "github.com/cortezaproject/corteza-server/system/types"
//	"github.com/stretchr/testify/require"
//	"testing"
//	"time"
//)
//
//func TestRecordValueSetter_Execute(t *testing.T) {
//	var (
//		scp molding.Variables
//		mod = &types.Module{
//			ID:     id.Next(),
//			Handle: "myModule",
//			Name:   "My Module",
//			Fields: types.ModuleFieldSet{
//				&types.ModuleField{Name: "fld1", Kind: "String"},
//				&types.ModuleField{Name: "fld2", Kind: "String"},
//				&types.ModuleField{Name: "num1", Kind: "Number"},
//			},
//			NamespaceID: id.Next(),
//			CreatedAt:   time.Time{},
//		}
//	)
//
//	t.Run("empty", func(t *testing.T) {
//		var (
//			req      = require.New(t)
//			rvs, err = NewRecordValueModifier("myRecord", mod, nil, nil)
//		)
//
//		req.NoError(err)
//		_, err = rvs.Execute(molding.Variables{})
//		req.NoError(err)
//	})
//
//	t.Run("record props", func(t *testing.T) {
//		var (
//			usr   = &sysTypes.User{ID: id.Next()}
//			rec   = &types.Record{Values: types.RecordValueSet{}}
//			scope = molding.Variables{
//				"ownIt":       true,
//				"currentUser": usr,
//				"myRecord":    rec,
//			}
//
//			req      = require.New(t)
//			rvs, err = NewRecordValueModifier(
//				"myRecord",
//				mod,
//				map[string]string{"ownedBy": "ownIt ? currentUser.ID : 0"}, nil,
//			)
//		)
//
//		req.NoError(err)
//		scp, err = rvs.Execute(scope)
//		req.NoError(err)
//		req.Equal(rec, scp["myRecord"])
//		req.Equal(rec.OwnedBy, usr.ID)
//
//		scope["ownIt"] = false
//		scp, err = rvs.Execute(scope)
//		req.NoError(err)
//		req.Zero(rec.OwnedBy)
//	})
//
//	t.Run("record values", func(t *testing.T) {
//		var (
//			usr   = &sysTypes.User{ID: id.Next()}
//			rec   = &types.Record{Values: types.RecordValueSet{}}
//			scope = molding.Variables{
//				"ownIt":       true,
//				"currentUser": usr,
//				"myRecord":    rec,
//			}
//
//			req      = require.New(t)
//			rvs, err = NewRecordValueModifier(
//				"myRecord",
//				mod,
//				nil,
//				map[string]string{"fld1": `"foo"`},
//			)
//		)
//
//		req.NoError(err)
//		scp, err = rvs.Execute(scope)
//		req.NoError(err)
//		req.NotNil(rec.Values.Get("fld1", 0))
//		req.Equal("foo", rec.Values.Get("fld1", 0).Value)
//	})
//
//	t.Run("record value from another value", func(t *testing.T) {
//		var (
//			usr   = &sysTypes.User{ID: id.Next()}
//			rec   = &types.Record{Values: types.RecordValueSet{}}
//			scope = molding.Variables{
//				"ownIt":       true,
//				"currentUser": usr,
//				"myRecord":    rec,
//			}
//
//			req      = require.New(t)
//			rvs, err = NewRecordValueModifier(
//				"myRecord",
//				mod,
//				nil,
//				map[string]string{"fld2": "myRecord.values.fld1"},
//			)
//		)
//
//		req.NoError(err)
//		scp, err = rvs.Execute(scope)
//		req.NoError(err)
//		req.Equal(rec, scp["myRecord"])
//	})
//
//	t.Run("invalid record value type", func(t *testing.T) {
//		var (
//			usr   = &sysTypes.User{ID: id.Next()}
//			rec   = &types.Record{Values: types.RecordValueSet{}}
//			scope = molding.Variables{
//				"ownIt":       true,
//				"currentUser": usr,
//				"myRecord":    rec,
//			}
//
//			req      = require.New(t)
//			rvs, err = NewRecordValueModifier(
//				"myRecord",
//				mod,
//				nil,
//				map[string]string{"num1": `"foo"`},
//			)
//		)
//
//		req.NoError(err)
//		scp, err = rvs.Execute(scope)
//		req.NoError(err)
//		req.Equal(rec, scp["myRecord"])
//	})
//}
