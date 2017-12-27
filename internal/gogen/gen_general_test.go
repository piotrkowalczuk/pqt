package gogen_test

import (
	"database/sql"
	"fmt"
	"go/format"
	"testing"
	"time"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
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

func TestGenerator_WhereClause(t *testing.T) {
	exp := func(kind, body string) string {
		res := `
type T1Criteria struct {`
		if kind != "none" {
			res += fmt.Sprintf(`
	Xyz                    %s`, kind)
		}
		res += `
	operator               string
	child, sibling, parent *T1Criteria
}`

		res += fmt.Sprintf(`
func T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if c.child == nil {
		return _T1CriteriaWhereClause(comp, c, id)
	}
	node := c
	sibling := false
	for {
		if !sibling {
			if node.child != nil {
				if node.parent != nil {
					comp.WriteString("(")
				}
				node = node.child
				continue
			} else {
				comp.Dirty = false
				comp.WriteString("(")
				if err := _T1CriteriaWhereClause(comp, node, id); err != nil {
					return err
				}
				comp.WriteString(")")
			}
		}
		if node.sibling != nil {
			sibling = false
			comp.WriteString(" ")
			comp.WriteString(node.parent.operator)
			comp.WriteString(" ")
			node = node.sibling
			continue
		}
		if node.parent != nil {
			sibling = true
			if node.parent.parent != nil {
				comp.WriteString(")")
			}
			node = node.parent
			continue
		}

		break
	}
	return nil
}

%s`, body)
		return res
	}
	cases := []struct {
		hint string
		col  *pqt.Column
		exp  string
	}{
		{
			hint: "sql-null-type",
			col:  pqt.NewColumn("xyz", pqt.TypeIntegerBig()),
			exp: exp("sql.NullInt64", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if c.Xyz.Valid {
		if comp.Dirty {
			comp.WriteString(" AND ")
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableT1ColumnXyz); err != nil {
			return err
		}
		if _, err := comp.WriteString("="); err != nil {
			return err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return err
		}
		comp.Add(c.Xyz)
		comp.Dirty = true
	}
	return nil
}`),
		},
		{
			hint: "pointer",
			col:  pqt.NewColumn("xyz", pqt.TypeInteger()),
			exp: exp("*int32", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if c.Xyz != nil {
		if comp.Dirty {
			comp.WriteString(" AND ")
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableT1ColumnXyz); err != nil {
			return err
		}
		if _, err := comp.WriteString("="); err != nil {
			return err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return err
		}
		comp.Add(c.Xyz)
		comp.Dirty = true
	}
	return nil
}`),
		},
		{
			hint: "non-pointer-time",
			col:  pqt.NewColumn("xyz", pqtgo.TypeCustom(time.Now(), time.Now(), time.Now())),
			exp: exp("time.Time", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if !c.Xyz.IsZero() {
		if comp.Dirty {
			comp.WriteString(" AND ")
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableT1ColumnXyz); err != nil {
			return err
		}
		if _, err := comp.WriteString("="); err != nil {
			return err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return err
		}
		comp.Add(c.Xyz)
		comp.Dirty = true
	}
	return nil
}`),
		},
		{
			hint: "nullable-time",
			col:  pqt.NewColumn("xyz", pqt.TypeTimestampTZ(), pqt.WithNotNull()),
			exp: exp("pq.NullTime", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if c.Xyz.Valid {
		if comp.Dirty {
			comp.WriteString(" AND ")
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableT1ColumnXyz); err != nil {
			return err
		}
		if _, err := comp.WriteString("="); err != nil {
			return err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return err
		}
		comp.Add(c.Xyz)
		comp.Dirty = true
	}
	return nil
}`),
		},
		{
			hint: "unknown",
			col:  pqt.NewColumn("xyz", pqtgo.TypeCustom(struct{}{}, struct{}{}, nil)),
			exp: exp("none", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
    			return nil
    		}`),
		},
		{
			hint: "dynamic",
			col: func() *pqt.Column {
				arg1 := pqt.NewColumn("x", pqt.TypeIntegerBig())
				arg2 := pqt.NewColumn("y", pqt.TypeIntegerBig())
				dyn := pqt.NewDynamicColumn("xyz", &pqt.Function{Name: "fn", Type: pqt.TypeIntegerBig(), Args: []*pqt.FunctionArg{
					{
						Name: "arg1",
						Type: pqt.TypeIntegerBig(),
					},
					{
						Name: "arg2",
						Type: pqt.TypeIntegerBig(),
					},
				}}, arg1, arg2)
				_ = pqt.NewTable("example").AddColumn(arg1).AddColumn(arg2).AddColumn(dyn)
				return dyn
			}(),
			exp: exp("sql.NullInt64", `
func _T1CriteriaWhereClause(comp *Composer, c *T1Criteria, id int) error {
	if c.Xyz.Valid {
		if comp.Dirty {
			comp.WriteString(" AND ")
		}
		if _, err := comp.WriteString("fn"); err != nil {
			return err
		}
		if _, err := comp.WriteString("("); err != nil {
			return err
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableExampleColumnX); err != nil {
			return err
		}
		if _, err := comp.WriteString(", "); err != nil {
			return err
		}
		if err := comp.WriteAlias(id); err != nil {
			return err
		}
		if _, err := comp.WriteString(TableExampleColumnY); err != nil {
			return err
		}
		if _, err := comp.WriteString(")"); err != nil {
			return err
		}
		if _, err := comp.WriteString("="); err != nil {
			return err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return err
		}
		comp.Add(c.Xyz)
		comp.Dirty = true
	}
	return nil
}`),
		},
	}

	for _, c := range cases {
		t.Run(c.hint, func(t *testing.T) {
			t1 := pqt.NewTable("t1").AddColumn(c.col)

			g := &gogen.Generator{}
			g.Criteria(t1)
			g.WhereClause(t1)
			assertOutput(t, g.Printer, c.exp)
		})
	}

}

func TestGenerator_JoinClause(t *testing.T) {
	g := &gogen.Generator{}
	g.JoinClause()
	if g.Printer.String() == "" {
		t.Error("output should not be empty")
	}
}

func TestGenerator_ScanRows(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").
		AddColumn(pqt.NewColumn("example", pqt.TypeText())).
		AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Entity(t2)
	g.ScanRows(t2)
	assertOutput(t, g.Printer, `
// T2Entity ...
type T2Entity struct {
	// Example ...
	Example sql.NullString
	// T1 ...
	T1 *T1Entity
}

// ScanT2Rows helps to scan rows straight to the slice of entities.
func ScanT2Rows(rows Rows) (entities []*T2Entity, err error) {
	for rows.Next() {
		var ent T2Entity
		err = rows.Scan(
			&ent.Example,
		)
		if err != nil {
			return
		}

		entities = append(entities, &ent)
	}
	if err = rows.Err(); err != nil {
		return
	}

	return
}`)
}

func TestGenerator_Statics(t *testing.T) {
	g := &gogen.Generator{}
	g.Statics(pqt.NewSchema("example"))
	_, err := format.Source(g.Bytes())
	if err != nil {
		t.Fatalf("unexpected printer formatting error: %s\n\n%s", err.Error(), g.String())
	}
}
