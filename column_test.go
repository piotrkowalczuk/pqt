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
	if d, ok := c.Default[pqt.EventInsert]; ok && d != "janusz" {
		t.Errorf("wrong column default, expected %s but got %s", "janusz", d)
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
		t.Errorf("wrong column reference, expected %p but got %p", r, c.Reference)
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

func TestColumn_DefaultOn(t *testing.T) {
	success := []struct {
		d string
		e []pqt.Event
	}{
		{
			d: "NOW()",
			e: []pqt.Event{pqt.EventUpdate},
		},
	}

	for _, data := range success {
		c := pqt.NewColumn("column", pqt.TypeTimestampTZ(), pqt.WithDefault(data.d, data.e...))

	EventLoop:
		for _, e := range data.e {
			d, ok := c.DefaultOn(e)
			if !ok {
				t.Errorf("missing default value for %s", e)
				continue EventLoop
			}

			if d != data.d {
				t.Errorf("wrong value, expected %s but got %s", data.d, d)
			}
		}
	}
}

func TestWithIndex(t *testing.T) {
	c := pqt.NewColumn("with_index", pqt.TypeText(), pqt.WithIndex())
	if !c.Index {
		t.Fatal("index expected to be true")
	}
}

func TestWithDefault(t *testing.T) {
	def := "0"
	c := pqt.NewColumn("with_index", pqt.TypeText(), pqt.WithDefault(def, pqt.EventInsert, pqt.EventUpdate))
	if len(c.Default) != 2 {
		t.Fatal("expected default value for 2 events")
	}
	if d, ok := c.Default[pqt.EventInsert]; ok {
		if d != def {
			t.Errorf("insert event wrong value, expected %s but got %s", def, d)
		}
	} else {
		t.Error("insert event expected")
	}
	if d, ok := c.Default[pqt.EventUpdate]; ok {
		if d != def {
			t.Errorf("update event wrong value, expected %s but got %s", def, d)
		}
	} else {
		t.Error("update event expected")
	}
}
