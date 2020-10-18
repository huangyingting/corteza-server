package action

//import (
//	"context"
//	"github.com/PaesslerAG/gval"
//	"github.com/cortezaproject/corteza-server/compose/types"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestFoo(t *testing.T) {
//	var (
//		ctx  = context.Background()
//		lang = gval.Full()
//
//		fields = types.ModuleFieldSet{
//			{Name: "foo", Kind: "String", Multi: true},
//			{Name: "bar", Kind: "Number", Multi: true},
//			{Name: "baz", Kind: "String"},
//			{Name: "is", Kind: "Bool"},
//		}
//
//		scope = map[string]interface{}{
//			"rec": recordValuesToMap(fields, types.RecordValueSet{
//				{Name: "foo", Value: "foo1"},
//				{Name: "foo", Value: "foo2"},
//				{Name: "foo", Value: "foo3"},
//				{Name: "bar", Value: "1"},
//				{Name: "bar", Value: "2"},
//				{Name: "baz", Value: "baz"},
//				{Name: "is", Value: "false"},
//			}),
//		}
//	)
//
//	ss := []struct {
//		name         string
//		expr         string
//		result       string
//		wantParseErr bool
//		wantEvalErr  bool
//	}{
//		{"access multi item", "rec.foo[0]", "foo1", false, false},
//		{"access all multi", "rec.foo", "[foo1 foo2 foo3]", false, false},
//		{"access string", "rec.baz", "baz", false, false},
//		{"concat strings", "rec.baz + rec.foo[1]", "bazfoo2", false, false},
//		{"math", "rec.bar[1] + 42", "44", false, false},
//		{"bool", "rec.is ? rec.baz : rec.foo[2]", "foo3", false, false},
//	}
//
//	for _, s := range ss {
//		t.Run(s.name, func(t *testing.T) {
//			var (
//				req       = require.New(t)
//				eval, err = lang.NewEvaluable(s.expr)
//			)
//
//			if s.wantParseErr {
//				req.Error(err)
//			} else {
//				req.NoError(err)
//			}
//
//			{
//				var result, err = eval.EvalString(ctx, scope)
//				if s.wantEvalErr {
//					req.Error(err)
//				} else {
//					req.NoError(err)
//				}
//
//				req.Equal(s.result, result)
//			}
//		})
//	}
//
//}
