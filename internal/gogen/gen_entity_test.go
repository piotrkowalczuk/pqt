package gogen_test

import (
	"fmt"
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_Entity(t *testing.T) {
	table := func(c *pqt.Column) *pqt.Table {
		t := pqt.NewTable("example")
		if c != nil {
			t.AddColumn(c)
		}
		return t
	}
	expected := func(columnName, columnType string) string {
		return fmt.Sprintf("\n// ExampleEntity ...\ntype ExampleEntity struct{\n// %s ...\nA %s}", columnName, columnType)
	}
	cases := map[string]struct {
		table *pqt.Table
		exp   string
	}{
		"simple": {
			table: table(nil),
			exp:   "\n// ExampleEntity ...\ntype ExampleEntity struct{}",
		},
		"column-bool": {
			table: table(pqt.NewColumn("a", pqt.TypeBool())),
			exp:   expected("A", "sql.NullBool"),
		},
		"column-bool-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeBool(), pqt.WithNotNull())),
			exp:   expected("A", "bool"),
		},
		"column-integer": {
			table: table(pqt.NewColumn("a", pqt.TypeInteger())),
			exp:   expected("A", "*int32"),
		},
		"column-integer-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeInteger(), pqt.WithNotNull())),
			exp:   expected("A", "int32"),
		},
		"column-integer-big": {
			table: table(pqt.NewColumn("a", pqt.TypeIntegerBig())),
			exp:   expected("A", "sql.NullInt64"),
		},
		"column-integer-big-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeIntegerBig(), pqt.WithNotNull())),
			exp:   expected("A", "int64"),
		},
		"dynamic": {
			table: func() *pqt.Table {
				age := pqt.NewColumn("age", pqt.TypeInteger())

				t := pqt.NewTable("example")
				t.AddColumn(age)
				t.AddColumn(pqt.NewDynamicColumn("dynamic", &pqt.Function{Type: pqt.TypeInteger()}, age))

				return t
			}(),
			exp: `
// ExampleEntity ...
type ExampleEntity struct{
// Age ...
Age *int32
// Dynamic ...
// Dynamic is read only
Dynamic int32}`,
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			g := &gogen.Generator{}
			g.Entity(c.table)
			assertOutput(t, g.Printer, c.exp)
		})
	}
}

func TestGenerator_EntityProp(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger())).
		AddColumn(pqt.NewColumn("doubles", pqt.TypeDoubleArray(0), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("integers_small", pqt.TypeIntegerSmallArray(0), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("integers_medium", pqt.TypeIntegerArray(0), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("integers_big", pqt.TypeIntegerBigArray(0), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("texts", pqt.TypeTextArray(0), pqt.WithNotNull()))

	g := &gogen.Generator{}
	g.Reset()
	g.Entity(t1)
	g.EntityProp(t1)
	assertOutput(t, g.Printer, `
// T1Entity ...
type T1Entity struct {
	// Age ...
	Age *int32
	// Doubles ...
	Doubles pq.Float64Array
	// ID ...
	ID int64
	// IntegersBig ...
	IntegersBig pq.Int64Array
	// IntegersMedium ...
	IntegersMedium pq.Int64Array
	// IntegersSmall ...
	IntegersSmall pq.Int64Array
	// Texts ...
	Texts pq.StringArray
}

func (e *T1Entity) Prop(cn string) (interface{}, bool) {
	switch cn {

	case TableT1ColumnAge:
		return e.Age, true
	case TableT1ColumnDoubles:
		if e.Doubles == nil {
			e.Doubles = []float64{}
		}
		return &e.Doubles, true
	case TableT1ColumnID:
		return &e.ID, true
	case TableT1ColumnIntegersBig:
		if e.IntegersBig == nil {
			e.IntegersBig = []int64{}
		}
		return &e.IntegersBig, true
	case TableT1ColumnIntegersMedium:
		if e.IntegersMedium == nil {
			e.IntegersMedium = []int64{}
		}
		return &e.IntegersMedium, true
	case TableT1ColumnIntegersSmall:
		if e.IntegersSmall == nil {
			e.IntegersSmall = []int64{}
		}
		return &e.IntegersSmall, true
	case TableT1ColumnTexts:
		if e.Texts == nil {
			e.Texts = []string{}
		}
		return &e.Texts, true
	default:
		return nil, false
	}
}`)
}
