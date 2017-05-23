package pqt_test

import (
	"reflect"
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestWithSchemaIfNotExists(t *testing.T) {
	sch := pqt.NewSchema("schema", pqt.WithSchemaIfNotExists())
	if !sch.IfNotExists {
		t.Error("expected if not exists to be true")
	}
}

func TestSchema_AddTable(t *testing.T) {
	tbl := pqt.NewTable("table")
	sch := pqt.NewSchema("schema").AddTable(tbl).AddTable(tbl)
	if !reflect.DeepEqual(sch, tbl.Schema) {
		t.Error("wrong schema assigned to the table")
	}
}

func TestSchema_AddFunction(t *testing.T) {
	fnc := pqt.FunctionNow()
	tbl := pqt.NewSchema("schema").AddFunction(fnc)
	if len(tbl.Functions) != 1 {
		t.Errorf("wrong number of functions: %d", len(tbl.Functions))
	}
}
