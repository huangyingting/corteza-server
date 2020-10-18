package action

//import (
//	"context"
//	"fmt"
//	"github.com/PaesslerAG/gval"
//	"github.com/cortezaproject/corteza-server/compose/types"
//	"github.com/cortezaproject/corteza-server/molding"
//	"github.com/cortezaproject/corteza-server/pkg/slice"
//	"github.com/spf13/cast"
//	"reflect"
//	"strconv"
//	"time"
//)
//
//type (
//	setRecordPairs map[string]gval.Evaluable
//
//	modifyRecordValue struct {
//		variable string
//		module   *types.Module
//		record   setRecordPairs
//		values   setRecordPairs
//	}
//)
//
//var (
//	_ molding.Executor = &modifyRecordValue{}
//)
//
//// NewRecordValueModifier handles setting variables for record and it's record values
////
//// @todo record prop setting should be split into a general variable-setting action
//func NewRecordValueModifier(variable string, mod *types.Module, r map[string]string, v map[string]string) (a *modifyRecordValue, err error) {
//	a = &modifyRecordValue{
//		variable: variable,
//		module:   mod,
//		record:   make(setRecordPairs),
//		values:   make(setRecordPairs),
//	}
//
//	var (
//		modifiableRecProps = slice.ToStringBoolMap([]string{
//			"ownedBy",
//			"createdAt",
//			"createdBy",
//			"updatedAt",
//			"updatedBy",
//			"deletedAt",
//			"deletedBy",
//		})
//
//		lang = gval.Full()
//	)
//
//	for key := range r {
//		if !modifiableRecProps[key] {
//			return nil, fmt.Errorf("unexisting or read-only record property %q used", key)
//		}
//	}
//
//	if a.module != nil {
//		for key := range v {
//			if mod.Fields.FindByName(key) == nil {
//				return nil, fmt.Errorf("unexisting field %q used", key)
//			}
//		}
//	}
//
//	for key, expr := range r {
//		if a.record[key], err = lang.NewEvaluable(expr); err != nil {
//			return nil, fmt.Errorf("can no evaluate record expression %q for %q: %w", expr, key, err)
//		}
//	}
//
//	for key, expr := range v {
//		if a.values[key], err = lang.NewEvaluable(expr); err != nil {
//			return nil, fmt.Errorf("can no evaluate record value expression %q for %q: %w", expr, key, err)
//		}
//	}
//
//	return a, nil
//}
//
//func (a *modifyRecordValue) Execute(scope molding.Variables) (molding.Variables, error) {
//	v, has := scope[a.variable]
//	if !has {
//		return scope, nil
//	}
//
//	rec, ok := v.(*types.Record)
//	if !ok {
//		return nil, fmt.Errorf("expecting type *Record for %q, got %T", a.variable, rec)
//	}
//
//	for prop, expr := range a.record {
//		rval, err := expr(context.Background(), scope)
//		if err != nil {
//			return nil, fmt.Errorf("failed to evaluate expression for record propery %q: %w", prop, err)
//		}
//
//		switch prop {
//		case "ownedBy":
//			rec.OwnedBy, ok = convertToUint64(rval)
//		case "createdAt":
//			if tmp, valid := convertToTime(rval); valid {
//				rec.CreatedAt = *tmp
//			} else {
//				ok = false
//			}
//		case "createdBy":
//			rec.CreatedBy, ok = convertToUint64(rval)
//		case "updatedAt":
//			rec.UpdatedAt, ok = convertToTime(rval)
//		case "updatedBy":
//			rec.UpdatedBy, ok = convertToUint64(rval)
//		case "deletedAt":
//			rec.DeletedAt, ok = convertToTime(rval)
//		case "deletedBy":
//			rec.DeletedBy, ok = convertToUint64(rval)
//		}
//
//		if !ok {
//			return nil, fmt.Errorf("failed to convert output value %v for record propery %q", rval, prop)
//		}
//	}
//
//	// Merge scope and convert *type.Record
//	var rvScope = molding.Variables{}
//	for k, v := range scope {
//		switch cnv := v.(type) {
//		case *types.Record:
//			// @todo where do we get the module from?!?
//			rvScope[k] = recordValuesToMap(nil, cnv.Values)
//		default:
//			rvScope[k] = v
//		}
//	}
//
//	for field, expr := range a.values {
//		var (
//			err error
//			// @todo handle multi-val!
//			val = rec.Values.Get(field, 0)
//		)
//
//		if val == nil {
//			val = &types.RecordValue{Name: field, Place: 0}
//		}
//
//		val.Value, err = expr.EvalString(context.Background(), rvScope)
//		if err != nil {
//			return nil, fmt.Errorf("failed to evaluate expression for record value %q: %w", field, err)
//		}
//
//		// @todo handle value validation!
//		rec.Values = rec.Values.Set(val)
//	}
//
//	return scope, nil
//}
//
//func (a *modifyRecordValue) Assign() map[string]string {
//	panic("implement me")
//}
//
//func (a *modifyRecordValue) Expects() molding.Params { return nil }
//
//func convertToUint64(o interface{}) (uint64, bool) {
//	if i, ok := o.(uint64); ok {
//		return i, true
//	}
//
//	if s, ok := o.(string); ok {
//		f, err := strconv.ParseUint(s, 10, 64)
//		if err == nil {
//			return f, true
//		}
//	}
//
//	v := reflect.ValueOf(o)
//	for o != nil && v.Kind() == reflect.Ptr {
//		v = v.Elem()
//		o = v.Interface()
//	}
//
//	switch v.Kind() {
//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//		return uint64(v.Int()), true
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
//		return v.Uint(), true
//	case reflect.Float32, reflect.Float64:
//		return uint64(v.Float()), true
//	}
//
//	return 0, false
//}
//
//func convertToTime(o interface{}) (*time.Time, bool) {
//	if i, ok := o.(time.Time); ok {
//		return &i, true
//	}
//
//	if i, ok := o.(*time.Time); ok {
//		return i, true
//	}
//
//	if s, ok := o.(string); ok {
//		t, err := cast.StringToDate(s)
//		if err == nil {
//			return &t, true
//		}
//	}
//
//	return nil, false
//}
