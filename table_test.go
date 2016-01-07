package pqt_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestTable_AddColumn(t *testing.T) {
	c1 := &pqt.Column{Name: "c1"}
	c2 := &pqt.Column{Name: "c2"}
	c3 := &pqt.Column{Name: "c3"}

	tbl := pqt.NewTable("test").
		AddColumn(c1).
		AddColumn(c2).
		AddColumn(c3)

	if len(tbl.Columns) != 3 {
		t.Errorf("wrong number of colums, expected %d but got %d", 3, len(tbl.Columns))
	}

	for i, c := range tbl.Columns {
		if c.Name == "" {
			t.Errorf("column #%d table name is empty", i)
		}
		if c.Table == nil {
			t.Errorf("column #%d table nil pointer", i)
		}
	}
}
