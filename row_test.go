package pqt_test

import (
	"testing"

	"go/types"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestNewColumn(t *testing.T) {
	collate := "UTF-7"
	check := "username = 'random'"
	r := pqt.NewColumn("username", pqt.TypeText())
	c := pqt.NewColumn(
		"user_username",
		pqt.TypeText(),
		pqt.WithCollate(collate),
		pqt.WithCheck(check),
		pqt.WithDefault("janusz"),
		pqt.WithUnique(),
		pqt.WithTypeMapping(pqtgo.BuiltinType(types.Byte)),
		pqt.WithNotNull(),
		pqt.WithPrimaryKey(),
		pqt.WithReference(r),
	)

	if c.Type.String() != pqt.TypeText().String() {
		t.Errorf("wrong column type, expected %s but got %s", pqt.TypeText().String(), c.Type.String())
	}
	if c.Collate != collate {
		t.Errorf("wrong column collate, expected %s but got %s", collate, c.Collate)
	}
	if c.Check != check {
		t.Errorf("wrong column check, expected %s but got %s", check, c.Check)
	}
	if c.Default != "janusz" {
		t.Errorf("wrong column default, expected %s but got %s", "janusz", c.Default)
	}
	if !c.Unique {
		t.Error("wrong column unique, expected true but got false")
	}
	if !c.NotNull {
		t.Error("wrong column not null, expected true but got false")
	}
	if !c.PrimaryKey {
		t.Error("wrong column primary key, expected true but got false")
	}
	if c.Reference != r {
		t.Errorf("wrong column reference, expected &p but got %p", r, c.Reference)
	}

	constraints := c.Constraints()

	if len(constraints) != 3 {
		t.Errorf("wrong number of constraints, expected 3 but got %d", len(constraints))
	}

	var hasPK, hasFK, hasCH bool
	for _, constraint := range constraints {
		switch constraint.Type {
		case pqt.ConstraintTypePrimaryKey:
			hasPK = true
		case pqt.ConstraintTypeForeignKey:
			hasFK = true
		case pqt.ConstraintTypeCheck:
			hasCH = true
		}
	}

	if !hasPK {
		t.Errorf("mising primary key constraint")
	}
	if !hasFK {
		t.Errorf("mising foreign key constraint")
	}
	if !hasCH {
		t.Errorf("mising check constraint")
	}
}
