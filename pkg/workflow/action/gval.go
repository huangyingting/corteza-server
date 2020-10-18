package action

//import (
//	"github.com/cortezaproject/corteza-server/compose/types"
//	"strconv"
//)
//
//// foo converts record values into map
////
//// @todo resolve references
//func recordValuesToMap(ff types.ModuleFieldSet, rvs types.RecordValueSet) map[string]interface{} {
//	var (
//		out = make(map[string]interface{})
//		val = func(f *types.ModuleField, r *types.RecordValue) interface{} {
//			switch true {
//			case f.IsBoolean():
//				b, _ := strconv.ParseBool(r.Value)
//				return b
//			default:
//				return r.Value
//			}
//		}
//	)
//
//	for _, f := range ff {
//		if f.Multi {
//			vv := rvs.FilterByName(f.Name)
//			mv := make([]interface{}, len(vv))
//			for i, rv := range rvs.FilterByName(f.Name) {
//				mv[i] = val(f, rv)
//			}
//
//			out[f.Name] = mv
//		} else {
//			out[f.Name] = val(f, rvs.Get(f.Name, 0))
//		}
//
//	}
//
//	return out
//}
