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

func TestGenerator_Package(t *testing.T) {
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

func TestGenerator_Imports(t *testing.T) {
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

func TestGenerator_Criteria(t *testing.T) {
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

func TestGenerator_Operand(t *testing.T) {
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

func TestGenerator_Repository(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Repository(t2)
	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)
}

func TestGenerator_Columns(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").
		AddColumn(pqt.NewColumn("id", pqt.TypeIntegerBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("name", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("description", pqt.TypeText())).
		AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Columns(t2)
	assertOutput(t, g.Printer, `
const (
	TableT2                  = "t2"
	TableT2ColumnDescription = "description"
	TableT2ColumnID          = "id"
	TableT2ColumnName        = "name"
)

var TableT2Columns = []string{
	TableT2ColumnDescription,
	TableT2ColumnID,
	TableT2ColumnName,
}`)
}

func TestGenerator_Constraints(t *testing.T) {
	name := pqt.NewColumn("name", pqt.TypeText(), pqt.WithNotNull(), pqt.WithIndex())
	description := pqt.NewColumn("description", pqt.TypeText(), pqt.WithColumnShortName("desc"))

	t1ID := pqt.NewColumn("id", pqt.TypeIntegerBig(), pqt.WithPrimaryKey())

	t1 := pqt.NewTable("t1").AddColumn(t1ID)
	t2 := pqt.NewTable("t2").
		AddColumn(pqt.NewColumn("id", pqt.TypeIntegerBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("t1_id", pqt.TypeIntegerBig(), pqt.WithReference(t1ID))).
		AddColumn(name).
		AddColumn(description).
		AddUnique(name, description).
		AddCheck("name <> 'LOL'", name)

	pqt.NewSchema("constraints_test").
		AddTable(t1).
		AddTable(t2)

	g := &gogen.Generator{}
	g.Constraints(t2)
	assertOutput(t, g.Printer, `
const (
	TableT2ConstraintPrimaryKey            = "constraints_test.t2_id_pkey"
	TableT2ConstraintT1IDForeignKey        = "constraints_test.t2_t1_id_fkey"
	TableT2ConstraintNameIndex             = "constraints_test.t2_name_idx"
	TableT2ConstraintNameDescriptionUnique = "constraints_test.t2_name_desc_key"
	TableT2ConstraintNameCheck             = "constraints_test.t2_name_check"
)`)
}

func TestGenerator_FindExpr(t *testing.T) {
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

func TestGenerator_CountExpr(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").
		AddRelationship(pqt.ManyToOne(t1)).
		AddColumn(pqt.NewColumn("name", pqt.TypeText()))

	g := &gogen.Generator{}
	g.CountExpr(t2)
	assertOutput(t, g.Printer, `
type T2CountExpr struct {
	Where  *T2Criteria
	JoinT1 *T1Join
}`)
}

func TestGenerator_Join(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Join(t2)
	assertOutput(t, g.Printer, `
type T2Join struct {
	On, Where *T2Criteria
	Fetch     bool
	Kind      JoinType
	JoinT1    *T1Join
}`)
}

func TestGenerator_Patch(t *testing.T) {
	table := func(columns ...*pqt.Column) *pqt.Table {
		t := pqt.NewTable("example")
		for _, c := range columns {
			t.AddColumn(c)

		}
		return t
	}
	expected := func(columns ...testColumn) string {
		res := fmt.Sprint(`
type ExamplePatch struct {`)
		for _, col := range columns {
			res += fmt.Sprintf(`
%s	%s`, col.name, col.kind)
		}
		return res + `}`
	}

	cases := map[string]struct {
		table *pqt.Table
		exp   string
	}{
		"primary_key": {
			table: table(
				pqt.NewColumn("a", pqt.TypeIntegerBig(), pqt.WithPrimaryKey()),
				pqt.NewColumn("b", pqt.TypeBool()),
			),
			exp: expected(testColumn{"B", "sql.NullBool"}),
		},
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
			g.Patch(c.table)
			assertOutput(t, g.Printer, c.exp)
		})
	}
}

func TestGenerator_Iterator(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Iterator(t2)
	assertOutput(t, g.Printer, `
// T2Iterator is not thread safe.
type T2Iterator struct {
	rows Rows
	cols []string
	expr *T2FindExpr
}

func (i *T2Iterator) Next() bool {
			return i.rows.Next()
		}

func (i *T2Iterator) Close() error {
	return i.rows.Close()
}

func (i *T2Iterator) Err() error {
	return i.rows.Err()
}

// Columns is wrapper around sql.Rows.Columns method, that also cache output inside iterator.
func (i *T2Iterator) Columns() ([]string, error) {
	if i.cols == nil {
		cols, err := i.rows.Columns()
		if err != nil {
			return nil, err
		}
		i.cols = cols
	}
	return i.cols, nil
}

// Ent is wrapper around T2 method that makes iterator more generic.
func (i *T2Iterator) Ent() (interface{}, error) {
	return i.T2()
}

func (i *T2Iterator) T2() (*T2Entity, error) {
	var ent T2Entity
	cols, err := i.Columns()
	if err != nil {
		return nil, err
	}

	props, err := ent.Props(cols...)
	if err != nil {
		return nil, err
	}
	var prop []interface{}
	if i.expr.JoinT1 != nil && i.expr.JoinT1.Fetch {
		ent.T1 = &T1Entity{}
		if prop, err = ent.T1.Props(); err != nil {
			return nil, err
		}
		props = append(props, prop...)
	}
	if err := i.rows.Scan(props...); err != nil {
		return nil, err
	}
	return &ent, nil
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
