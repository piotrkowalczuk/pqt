package pqt_test

import (
	"testing"

	"go/types"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestNewDynamicColumn(t *testing.T) {
	c := pqt.NewDynamicColumn("dynamic", &pqt.Function{
		Name:      "sum",
		Type:      pqt.TypeIntegerBig(),
		Body:      "select $1 + $2",
		Behaviour: pqt.FunctionBehaviourStable,
		Args: []*pqt.FunctionArg{
			{
				Name: "A",
				Type: pqt.TypeIntegerBig(),
			},
			{
				Name: "B",
				Type: pqt.TypeIntegerBig(),
			},
		},
	})
	if !c.IsDynamic {
		t.Error("column should be dynamic")
	}
	if c.Name != "dynamic" {
		t.Errorf("wrong name: %s", c.Name)
	}
	if len(c.Func.Args) != 2 {
		t.Errorf("wrong number of function arguments, expected %d got %d", 2, len(c.Func.Args))
	}
}

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

func TestWithOnDelete(t *testing.T) {
	c := pqt.NewColumn("on_delete", pqt.TypeBool(), pqt.WithOnDelete(pqt.Cascade))
	if c.OnDelete != pqt.Cascade {
		t.Errorf("wrong on delete event: %s", c.OnDelete)
	}
}

func TestWithOnUpdate(t *testing.T) {
	c := pqt.NewColumn("on_update", pqt.TypeBool(), pqt.WithOnUpdate(pqt.SetNull))
	if c.OnUpdate != pqt.SetNull {
		t.Errorf("wrong on update event: %s", c.OnUpdate)
	}
}

func TestWithColumnShortName(t *testing.T) {
	given := "short-name"
	c := pqt.NewColumn("short_name", pqt.TypeBool(), pqt.WithColumnShortName(given))
	if c.ShortName != given {
		t.Errorf("wrong short name: %s", c.ShortName)
	}
}

func TestColumns_String(t *testing.T) {
	given := pqt.Columns{
		&pqt.Column{Name: "1"},
		&pqt.Column{Name: "2"},
		&pqt.Column{Name: "3"},
	}
	exp := "1,2,3"
	if given.String() != exp {
		t.Errorf("wrong output, expected %s but got %s", exp, given.String())
	}
}

func TestJoinColumns(t *testing.T) {
	got := pqt.JoinColumns(pqt.Columns{
		&pqt.Column{Name: "1"},
		&pqt.Column{Name: "2"},
		&pqt.Column{Name: "3"},
	}, ".")
	exp := "1.2.3"
	if got != exp {
		t.Errorf("wrong output, expected %s but got %s", exp, got)
	}
}

func TestColumn_Constraints(t *testing.T) {
	check := "column > 0"
	col := pqt.NewColumn(
		"column",
		pqt.TypeSerial(),
		pqt.WithPrimaryKey(),
		pqt.WithUnique(),
		pqt.WithCheck(check),
		pqt.WithIndex(),
	)
	tbl := pqt.NewTable("table").AddColumn(col)
	got := col.Constraints()

	var nb int
	for _, g := range got {
		switch g.Type {
		case pqt.ConstraintTypePrimaryKey:
			nb++
			if len(g.Columns) != 1 {
				t.Errorf("wrong number of columns, expected 1 got %d", len(g.Columns))
			}
			if g.Table != tbl {
				t.Error("wrong table")
			}
		case pqt.ConstraintTypeIndex, pqt.ConstraintTypeUnique:
			t.Errorf("unexpected constraint type, if column has primary key index and unique should be ignored, got %s", g.Type)
		case pqt.ConstraintTypeCheck:
			nb++
			if len(g.Columns) != 1 {
				t.Errorf("pk: wrong number of columns, expected 1 got %d", len(g.Columns))
			}
			if g.Table != tbl {
				t.Error("wrong table")
			}
			if g.Check != check {
				t.Error("wrong check")
			}
		case pqt.ConstraintTypeForeignKey:
			nb++
		}
	}
	if nb != 2 {
		t.Errorf("wrong number of constraints, expected 2 got %d", nb)
	}

	col = pqt.NewColumn(
		"column",
		pqt.TypeSerial(),
		pqt.WithUnique(),
		pqt.WithIndex(),
	)
	nb = 0
	tbl = pqt.NewTable("table").AddColumn(col)

	for _, g := range col.Constraints() {
		switch g.Type {
		case pqt.ConstraintTypeUnique:
			nb++
			if len(g.Columns) != 1 {
				t.Errorf("wrong number of columns, expected 1 got %d", len(g.Columns))
			}
			if g.Table != tbl {
				t.Error("wrong table")
			}
		case pqt.ConstraintTypeIndex:
			t.Errorf("unexpected constraint type, if column has unique index regular index be ignored, got %s", g.Type)
		}
	}
	if nb != 1 {
		t.Errorf("wrong number of constraints, expected 1 got %d", nb)
	}

	col = pqt.NewColumn(
		"column",
		pqt.TypeSerial(),
		pqt.WithIndex(),
	)
	nb = 0
	tbl = pqt.NewTable("table").AddColumn(col)

	for _, g := range col.Constraints() {
		switch g.Type {
		case pqt.ConstraintTypeIndex:
			nb++
			if len(g.Columns) != 1 {
				t.Errorf("wrong number of columns, expected 1 got %d", len(g.Columns))
			}
			if g.Table != tbl {
				t.Error("wrong table")
			}
		}
	}
}
