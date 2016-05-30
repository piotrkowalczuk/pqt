package pqt_test

import (
	"testing"

	"reflect"

	"github.com/piotrkowalczuk/pqt"
)

func TestConstraint_Name(t *testing.T) {
	id := pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey())
	success := map[string]*pqt.Constraint{
		"public.user_id_pkey": pqt.PrimaryKey(pqt.NewTable("user"), id),
		"custom_schema.user_id_pkey": pqt.PrimaryKey(func() *pqt.Table {
			t := pqt.NewTable("user")
			s := pqt.NewSchema("custom_schema")
			s.AddTable(t)

			return t
		}(), id),
		"<missing table>": pqt.Check(nil, "a > b", id),
		"public.news_key": pqt.Unique(pqt.NewTable("news")),
	}

	for expected, given := range success {
		got := given.Name()

		if got != expected {
			t.Errorf("wrong name, expected %s got %s", expected, got)
		}
	}
}

func TestForeignKey(t *testing.T) {
	t1 := pqt.NewTable("left")
	c11 := pqt.NewColumn("id", pqt.TypeSerialBig())
	c12 := pqt.NewColumn("name", pqt.TypeText())
	t1.AddColumn(c11)
	t1.AddColumn(c12)

	t2 := pqt.NewTable("right")
	c21 := pqt.NewColumn("id", pqt.TypeSerialBig())
	c22 := pqt.NewColumn("name", pqt.TypeText())
	t2.AddColumn(c21)
	t2.AddColumn(c22)

	cstr := pqt.ForeignKey(t1, pqt.Columns{c11, c12}, pqt.Columns{c21, c22})
	if cstr.Type != pqt.ConstraintTypeForeignKey {
		t.Errorf("wrong type, expected %s but got %s", pqt.ConstraintTypeForeignKey, cstr.Type)
	}
	if len(cstr.Columns) != 2 {
		t.Errorf("wrong number of columns, expected %d but got %d", 2, len(cstr.Columns))
	}
	if len(cstr.ReferenceColumns) != 2 {
		t.Errorf("wrong number of columns, expected %d but got %d", 2, len(cstr.ReferenceColumns))
	}
	if !reflect.DeepEqual(cstr.Table, t1) {
		t.Errorf("table does not match, expected %v but got %v", t1, cstr.Table)
	}
	if !reflect.DeepEqual(cstr.ReferenceTable, t2) {
		t.Errorf("reference table does not match, expected %v but got %v", t2, cstr.ReferenceTable)
	}
}
