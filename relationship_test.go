package pqt_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestWithInversedColumnName(t *testing.T) {
	icn := "author"
	t1 := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	t2 := pqt.NewTable("comment")

	t1.AddRelationship(pqt.OneToOneUnidirectional(t2, pqt.WithInversedColumnName(icn)))

	if len(t1.Relationships) != 0 {
		t.Fatalf("user table should have exactly 0 relationship, got %d", len(t1.Relationships))
	}

	if len(t2.Relationships) != 1 {
		t.Fatalf("comment table should have exactly 1 relationship, got %d", len(t2.Relationships))
	}

	var exists bool
	for _, c := range t2.Columns {
		if c.Name == icn {
			exists = true
			break
		}
	}

	if !exists {
		t.Errorf("comment table should have collumn with name %s", icn)
	}
}
