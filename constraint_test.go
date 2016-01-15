package pqt_test

import (
	"testing"

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
