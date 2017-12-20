package gogen_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/format"
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestPackage(t *testing.T) {
	cases := []struct {
		name, exp string
	}{
		{
			name: "something",
			exp:  "package something\n",
		},
		{
			name: "",
			exp:  "package main\n",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			g := &gogen.Generator{}
			g.Package(c.name)
			assertOutput(t, g.Printer, c.exp)
		})
	}
}

func TestImports(t *testing.T) {
	cases := []struct {
		schema pqt.Schema
		fixed  []string
		exp    string
	}{
		{
			schema: pqt.Schema{},
			exp:    "import(\n\"github.com/m4rw3r/uuid\"\n)\n",
		},
		{
			schema: pqt.Schema{
				Tables: []*pqt.Table{
					{
						Columns: pqt.Columns{
							pqt.NewColumn("custom", pqtgo.TypeCustom("", sql.NullString{}, sql.NullString{})),
						},
					},
				},
			},
			exp: "import(\n\"github.com/m4rw3r/uuid\"\n\"database/sql\"\n\"database/sql\"\n)\n",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			g := &gogen.Generator{}
			g.Imports(&c.schema, c.fixed...)
			assertOutput(t, g.Printer, c.exp)
		})
	}
}

func TestEntity(t *testing.T) {
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
		fixed []string
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

func TestCriteria(t *testing.T) {
	table := func(c *pqt.Column) *pqt.Table {
		t := pqt.NewTable("example")
		if c != nil {
			t.AddColumn(c)
		}
		return t
	}
	expected := func(columns ...testColumn) string {
		res := fmt.Sprint(`
type ExampleCriteria struct {`)
		for _, col := range columns {
			res += fmt.Sprintf(`
%s	%s`, col.name, col.kind)
		}
		return res + fmt.Sprint(`
operator               string
child, sibling, parent *ExampleCriteria
}`)
	}

	cases := map[string]struct {
		table *pqt.Table
		fixed []string
		exp   string
	}{
		"column-bool": {
			table: table(pqt.NewColumn("a", pqt.TypeBool())),
			exp:   expected(testColumn{"A", "sql.NullBool"}),
		},
		"column-bool-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeBool(), pqt.WithNotNull())),
			exp:   expected(testColumn{"A", "sql.NullBool"}),
		},
		"column-integer": {
			table: table(pqt.NewColumn("a", pqt.TypeInteger())),
			exp:   expected(testColumn{"A", "*int32"}),
		},
		"column-integer-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeInteger(), pqt.WithNotNull())),
			exp:   expected(testColumn{"A", "*int32"}),
		},
		"column-integer-big": {
			table: table(pqt.NewColumn("a", pqt.TypeIntegerBig())),
			exp:   expected(testColumn{"A", "sql.NullInt64"}),
		},
		"column-integer-big-not-null": {
			table: table(pqt.NewColumn("a", pqt.TypeIntegerBig(), pqt.WithNotNull())),
			exp:   expected(testColumn{"A", "sql.NullInt64"}),
		},
		"dynamic": {
			table: func() *pqt.Table {
				age := pqt.NewColumn("age", pqt.TypeInteger())

				t := pqt.NewTable("example")
				t.AddColumn(age)
				t.AddColumn(pqt.NewDynamicColumn("dynamic", &pqt.Function{Type: pqt.TypeInteger()}, age))

				return t
			}(),
			exp: expected(testColumn{"Age", "*int32"}, testColumn{"Dynamic", "*int32"}),
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			g := &gogen.Generator{}
			g.Criteria(c.table)
			assertOutput(t, g.Printer, c.exp)
		})
	}
}

func TestOperand(t *testing.T) {
	g := &gogen.Generator{}
	g.Operand(pqt.NewTable("example"))
	assertOutput(t, g.Printer, `
func ExampleOperand(operator string, operands ...*ExampleCriteria) *ExampleCriteria {
	if len(operands) == 0 {
		return &ExampleCriteria{operator: operator}
	}

	parent := &ExampleCriteria{
		operator: operator,
		child:    operands[0],
	}

	for i := 0; i < len(operands); i++ {
		if i < len(operands)-1 {
			operands[i].sibling = operands[i+1]
		}
		operands[i].parent = parent
	}

	return parent
}

func ExampleOr(operands ...*ExampleCriteria) *ExampleCriteria {
	return ExampleOperand("OR", operands...)
}

func ExampleAnd(operands ...*ExampleCriteria) *ExampleCriteria {
	return ExampleOperand("AND", operands...)
}`)
}

func TestFindExpr(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.FindExpr(t2)
	assertOutput(t, g.Printer, `
type T2FindExpr struct {
	Where         *T2Criteria
	Offset, Limit int64
	Columns       []string
	OrderBy       []RowOrder
	JoinT1        *T1Join
}`)
}

type testColumn struct {
	name, kind string
}

func assertOutput(t *testing.T, p print.Printer, e string) {
	t.Helper()

	got, err := format.Source(p.Bytes())
	if err != nil {
		t.Fatalf("unexpected printer formatting error: %s", err.Error())
	}
	exp, err := format.Source([]byte(e))
	if err != nil {
		t.Fatalf("unexpected expected formatting error: %s", err.Error())
	}
	if !bytes.Equal(got, exp) {
		t.Errorf("wrong output, expected:\n'%s'\nbut got:\n'%s'", string(exp), string(got))
	}
}
