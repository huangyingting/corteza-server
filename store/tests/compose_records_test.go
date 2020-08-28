package tests

import (
	"context"
	"github.com/cortezaproject/corteza-server/compose/types"
	"github.com/cortezaproject/corteza-server/pkg/id"
	"github.com/cortezaproject/corteza-server/pkg/rh"
	"github.com/cortezaproject/corteza-server/store"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func testComposeRecords(t *testing.T, s store.ComposeRecords) {
	var (
		ctx = context.Background()

		mod = &types.Module{
			ID:          id.Next(),
			NamespaceID: id.Next(),
			Handle:      "",
			Name:        "testComposeRecords",
			CreatedAt:   time.Now(),
			Fields: types.ModuleFieldSet{
				&types.ModuleField{Kind: "String", Name: "str1"},
				&types.ModuleField{Kind: "String", Name: "str2"},
				&types.ModuleField{Kind: "String", Name: "str3"},
				&types.ModuleField{Kind: "Number", Name: "num1"},
				&types.ModuleField{Kind: "Number", Name: "num2"},
				&types.ModuleField{Kind: "Number", Name: "num3"},
				&types.ModuleField{Kind: "DateTime", Name: "dt1"},
			},
		}

		makeNew = func(vv ...*types.RecordValue) *types.Record {
			// minimum data set for new composeRecord
			var recordID = id.Next()

			for _, v := range vv {
				v.RecordID = recordID
			}

			return &types.Record{
				ID:          recordID,
				NamespaceID: mod.NamespaceID,
				ModuleID:    mod.ID,
				CreatedAt:   time.Now(),
				Values:      vv,
			}
		}

		truncAndCreate = func(t *testing.T, rr ...*types.Record) (*require.Assertions, types.RecordSet) {
			req := require.New(t)
			req.NoError(s.TruncateComposeRecords(ctx, mod))

			if len(rr) == 0 {
				rr = []*types.Record{makeNew()}
			}

			for _, rec := range rr {
				req.NoError(s.CreateComposeRecord(ctx, mod, rec))
			}

			return req, rr
		}
	)

	t.Run("create", func(t *testing.T) {
		req := require.New(t)
		composeRecord := makeNew()
		req.NoError(s.CreateComposeRecord(ctx, mod, composeRecord))
	})

	t.Run("lookup by ID", func(t *testing.T) {
		req, rr := truncAndCreate(t, makeNew(
			&types.RecordValue{Name: "str1", Value: "v1", Ref: 1},
			&types.RecordValue{Name: "str2", Value: "v2", Ref: 2},
			&types.RecordValue{Name: "str3", Value: "v3", Ref: 3},
		))
		rec := rr[0]

		fetched, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)
		req.NoError(err)
		req.Equal(rec.ID, fetched.ID)
		req.NotNil(fetched.CreatedAt)
		req.Nil(fetched.UpdatedAt)
		req.Nil(fetched.DeletedAt)
		req.Len(fetched.Values, len(rec.Values))
		req.Equal("str2", fetched.Values[1].Name)
		req.Equal("v2", fetched.Values[1].Value)
		req.Equal(uint64(2), fetched.Values[1].Ref)
	})

	t.Run("Delete", func(t *testing.T) {
		req, rr := truncAndCreate(t)
		rec := rr[0]

		req.NoError(s.DeleteComposeRecord(ctx, mod, rec))
		_, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)
		req.EqualError(err, store.ErrNotFound.Error())
	})

	t.Run("Delete by ID", func(t *testing.T) {
		req, rr := truncAndCreate(t)
		rec := rr[0]

		req.NoError(s.DeleteComposeRecordByID(ctx, mod, rec.ID))
		_, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)
		req.EqualError(err, store.ErrNotFound.Error())
	})

	t.Run("update", func(t *testing.T) {
		req, rr := truncAndCreate(t)
		rec := rr[0]

		rec = &types.Record{
			ID:          rec.ID,
			CreatedAt:   rec.CreatedAt,
			ModuleID:    mod.ID,
			NamespaceID: mod.NamespaceID,
			OwnedBy:     id.Next(),
		}

		req.NoError(s.UpdateComposeRecord(ctx, mod, rec))

		updated, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)
		req.NoError(err)
		req.Equal(rec.OwnedBy, updated.OwnedBy)
	})

	t.Run("update values", func(t *testing.T) {
		req, rr := truncAndCreate(t, makeNew(
			&types.RecordValue{Name: "str1", Value: "v1", Ref: 1},
			&types.RecordValue{Name: "str2", Value: "v2", Ref: 2},
		))
		rec := rr[0]

		rec = &types.Record{
			ID:          rec.ID,
			CreatedAt:   rec.CreatedAt,
			OwnedBy:     id.Next(),
			Values:      rec.Values,
			ModuleID:    mod.ID,
			NamespaceID: mod.NamespaceID,
		}

		rec.Values[0].Value = "vv10"
		rec.Values[1].Value = "vv20"
		rec.Values = append(rec.Values, &types.RecordValue{Name: "str3", Value: "vv30", Ref: 3})
		rec.Values.SetRecordID(rec.ID)

		req.NoError(s.UpdateComposeRecord(ctx, mod, rec))

		updated, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)
		req.NoError(err)
		req.Equal(rec.OwnedBy, updated.OwnedBy)
		req.Len(updated.Values, len(rec.Values))
		req.Equal("str2", updated.Values[1].Name)
		req.Equal("vv20", updated.Values[1].Value)
	})

	t.Run("soft delete values", func(t *testing.T) {
		req, rr := truncAndCreate(t, makeNew(
			&types.RecordValue{Name: "str1", Value: "v1", Ref: 1},
			&types.RecordValue{Name: "str2", Value: "v2", Ref: 2},
		))
		rec := rr[0]
		rec.DeletedAt = &rec.CreatedAt

		req.NoError(s.UpdateComposeRecord(ctx, mod, rec))

		updated, err := s.LookupComposeRecordByID(ctx, mod, rec.ID)

		req.NoError(err)
		req.NotNil(rec)
		req.NotNil(rec.DeletedAt)
		req.Len(updated.Values, len(rec.Values))
		req.NotNil(updated.Values[0].DeletedAt)
		req.NotNil(updated.Values[1].DeletedAt)
	})

	t.Run("search", func(t *testing.T) {
		t.Run("by record attributes", func(t *testing.T) {
			prefill := []*types.Record{
				makeNew(),
				makeNew(),
				makeNew(),
				makeNew(),
				makeNew(),
			}

			count := len(prefill)

			prefill[4].DeletedAt = &prefill[4].CreatedAt
			valid := count - 1

			req, _ := truncAndCreate(t, prefill...)

			// search for all valid
			set, _, err := s.SearchComposeRecords(ctx, mod, types.RecordFilter{})
			req.NoError(err)
			req.Len(set, valid) // we've deleted one

			// search for ALL
			set, _, err = s.SearchComposeRecords(ctx, mod, types.RecordFilter{Deleted: rh.FilterStateInclusive})
			req.NoError(err)
			req.Len(set, count) // we've deleted one

			// search for deleted only
			set, _, err = s.SearchComposeRecords(ctx, mod, types.RecordFilter{Deleted: rh.FilterStateExclusive})
			req.NoError(err)
			req.Len(set, 1) // we've deleted one
		})

		t.Run("by values", func(t *testing.T) {
			var (
				err error
				set types.RecordSet

				req, _ = truncAndCreate(t,
					makeNew(&types.RecordValue{Name: "str1", Value: "v1"}, &types.RecordValue{Name: "str2", Value: "same"}, &types.RecordValue{Name: "str3", Value: "three"}),
					makeNew(&types.RecordValue{Name: "str1", Value: "v2"}, &types.RecordValue{Name: "str2", Value: "same"}, &types.RecordValue{Name: "str3", Value: "three"}),
					makeNew(&types.RecordValue{Name: "str1", Value: "v3"}, &types.RecordValue{Name: "str2", Value: "same"}, &types.RecordValue{Name: "str3", Value: "three"}),
					makeNew(&types.RecordValue{Name: "str1", Value: "v4"}, &types.RecordValue{Name: "str2", Value: "same"}),
					makeNew(&types.RecordValue{Name: "str1", Value: "v5"}, &types.RecordValue{Name: "str2", Value: "same"}),

					// Add one additional record with deleted values
					makeNew(&types.RecordValue{Name: "str1", Value: "v6", DeletedAt: now()}, &types.RecordValue{Name: "str2", Value: "deleted", DeletedAt: now()}),
				)

				f = types.RecordFilter{
					ModuleID:    mod.ID,
					NamespaceID: mod.NamespaceID,
				}
			)

			f.Query = `str1 = 'v1'`
			set, _, err = s.SearchComposeRecords(ctx, mod, f)
			req.NoError(err)
			req.Len(set, 1)

			f.Query = `str2 = 'same'`
			set, _, err = s.SearchComposeRecords(ctx, mod, f)
			req.NoError(err)
			req.Len(set, 5)

			f.Query = `str2 = 'different'`
			set, _, err = s.SearchComposeRecords(ctx, mod, f)
			req.NoError(err)
			req.Len(set, 0)

			f.Query = `str3 = 'three' AND str1 = 'v1'`
			set, _, err = s.SearchComposeRecords(ctx, mod, f)
			req.NoError(err)
			req.Len(set, 1)
		})
	})

	t.Run("report", func(t *testing.T) {
		var (
			err error

			req, _ = truncAndCreate(t,
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-01-01"}, &types.RecordValue{Name: "num1", Value: "1"}, &types.RecordValue{Name: "str3", Value: "three"}),
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-01-01"}, &types.RecordValue{Name: "num1", Value: "2"}, &types.RecordValue{Name: "str3", Value: "three"}),
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-01-01"}, &types.RecordValue{Name: "num1", Value: "3"}, &types.RecordValue{Name: "str3", Value: "three"}),
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-05-01"}, &types.RecordValue{Name: "num1", Value: "4"}),
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-05-01"}, &types.RecordValue{Name: "num1", Value: "5"}),

				// Add one additional record with deleted values
				makeNew(&types.RecordValue{Name: "dt1", Value: "2020-05-01", DeletedAt: now()}, &types.RecordValue{Name: "num1", Value: "6", DeletedAt: now()}, &types.RecordValue{Name: "str2", Value: "deleted", DeletedAt: now()}),
			)

			report []map[string]interface{}
		)

		report, err = s.ComposeRecordReport(ctx, mod, "MAX(num1)", "QUARTER(dt1)", "")
		req.NoError(err)
		req.Len(report, 3)

		// @todo find a way to compare the results

		//expected := []map[string]interface{}{
		//	{"count": 3, "dimension_0": 1, "metric_0": 3},
		//	{"count": 2, "dimension_0": 2, "metric_0": 5},
		//	{"count": 1, "dimension_0": nil, "metric_0": nil},
		//}
		//
		////spew.Dump(report, expected)
		//req.True(
		//	reflect.DeepEqual(report, expected),
		//	"report does not match expected results:\n%#v\n%#v", report, expected)

		report, err = s.ComposeRecordReport(ctx, mod, "COUNT(num1)", "YEAR(dt1)", "")
		req.NoError(err)

		report, err = s.ComposeRecordReport(ctx, mod, "SUM(num1)", "DATE(dt1)", "")
		req.NoError(err)

		report, err = s.ComposeRecordReport(ctx, mod, "MIN(num1)", "DATE(NOW())", "")
		req.NoError(err)

		report, err = s.ComposeRecordReport(ctx, mod, "AVG(num1)", "DATE(NOW())", "")
		req.NoError(err)

		// Note that not all functions are compatible across all backends

	})

}