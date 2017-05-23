package model_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"strconv"

	"math"

	"reflect"

	"github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var testCompleteInsertData = map[string]struct {
	entity model.CompleteEntity
	query  string
}{
	"none": {
		entity: model.CompleteEntity{},
		query:  "INSERT INTO example.complete",
	},
	"full": {
		entity: model.CompleteEntity{
			ColumnBool: sql.NullBool{
				Valid: true,
				Bool:  true,
			},
			ColumnBytea: []byte("something"),
			ColumnCharacter0: sql.NullString{
				Valid:  true,
				String: "something 0",
			},
			ColumnCharacter100: sql.NullString{
				Valid:  true,
				String: "something 100",
			},
			ColumnDecimal: sql.NullFloat64{
				Valid:   true,
				Float64: 12.12,
			},
			ColumnDoubleArray0: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{11.11, 12.12},
			},
			ColumnDoubleArray100: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{11.11, 12.12},
			},
			ColumnInteger: func() *int32 {
				x := int32(1)
				return &x
			}(),
			ColumnIntegerArray0: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnIntegerArray100: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnIntegerBig: sql.NullInt64{
				Valid: true,
				Int64: math.MaxInt64,
			},
			ColumnIntegerBigArray0: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnIntegerBigArray100: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnIntegerSmall: func() *int16 {
				x := int16(1)
				return &x
			}(),
			ColumnIntegerSmallArray0: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnIntegerSmallArray100: model.NullInt64Array{
				Valid:      true,
				Int64Array: []int64{1, 2, 3},
			},
			ColumnJson:     []byte(`{"field": "null"}`),
			ColumnJsonNn:   []byte(`{"field": "not null"}`),
			ColumnJsonNnD:  []byte(`{"field": "not null, default"}`),
			ColumnJsonb:    []byte(`{"field": "value"}`),
			ColumnJsonbNn:  []byte(`{"field": "not null"}`),
			ColumnJsonbNnD: []byte(`{"field": "not null, default"}`),
			ColumnNumeric: sql.NullFloat64{
				Valid:   true,
				Float64: math.MaxFloat64,
			},
			ColumnReal: func() *float32 {
				x := float32(math.MaxFloat32)
				return &x
			}(),
			ColumnSerial: func() *int32 {
				x := int32(math.MaxInt32)
				return &x
			}(),
			ColumnSerialBig: sql.NullInt64{
				Valid: true,
				Int64: math.MaxInt64,
			},
			ColumnSerialSmall: func() *int16 {
				x := int16(math.MaxInt16)
				return &x
			}(),
			ColumnText: sql.NullString{
				Valid:  true,
				String: "some string",
			},
			ColumnTextArray0: model.NullStringArray{
				Valid:       true,
				StringArray: []string{"something 0"},
			},
			ColumnTextArray100: model.NullStringArray{
				Valid:       true,
				StringArray: []string{"something 100"},
			},
			ColumnTimestamp: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			ColumnTimestamptz: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			ColumnUUID: sql.NullString{
				Valid:  true,
				String: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a1111",
			},
		},
		query: "INSERT INTO example.complete (" + columns(model.TableCompleteColumns, "column_serial", "column_serial_big", "column_serial_small") + ") VALUES (" + func(args []string) (r string) {
			for i := range args {
				if i != 0 {
					r += ", "
				}
				r += "$" + strconv.FormatInt(int64(i)+1, 10)
			}
			return
		}(without(model.TableCompleteColumns, "column_serial", "column_serial_big", "column_serial_small")) + ") RETURNING " + strings.Join(model.TableCompleteColumns, ", "),
	},
}

func BenchmarkCompleteRepositoryBase_InsertQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testCompleteInsertData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.complete.InsertQuery(&given.entity, true)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestCompleteRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testCompleteInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.complete.InsertQuery(&given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCompleteRepositoryBase_Insert(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			got, err := s.news.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			given.entity.ID = got.ID
			given.entity.CreatedAt = got.CreatedAt

			if !reflect.DeepEqual(given.entity, *got) {
				t.Errorf("unequal entities, expected:\n	%v\nbut got:\n	%v", given.entity, *got)
			}
		})
	}
}

func without(cs []string, wo ...string) []string {
	var res []string
ColumnsLoop:
	for _, c := range cs {
		for _, w := range wo {
			if c == w {
				continue ColumnsLoop
			}
		}
		res = append(res, c)
	}
	return res
}

func columns(cs []string, wo ...string) string {
	return strings.Join(without(cs, wo...), ", ")
}
