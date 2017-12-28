package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) TestRepositoryInsertQuery(t *pqt.Table) {
	name := formatter.Public(t.Name)
	g.Printf(`
func Test%sRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range test%sInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.Get%s().InsertQuery(given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %%s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%%s\nbut got:\n	%%s", given.query, query)
			}
		})
	}
}`, name, name, name)
}

func (g *Generator) TestRepositoryInsert(t *pqt.Table) {
	name := formatter.Public(t.Name)
	g.Printf(`
func Test%sRepositoryBase_Insert(t *testing.T) {
	test%sRepositoryBaseInsert(t, 1000)
}`, name, name)
	g.Printf(`

func test%sRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)`, name)
	for _, rel := range joinableRelationships(t) {
		if rel.OwnerTable != t {
			continue
		}
		g.Printf(`
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.Get%s().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %%s\nfor entity: %%#v", err.Error(), ent)
			}
			assert%sEntity(t, ent, got)
		})
	}
}`, formatter.Public(rel.InversedTable.Name), formatter.Public(rel.InversedTable.Name))
	}
	for _, con := range t.Constraints {
		switch con.Type {
		case pqt.ConstraintTypeUnique, pqt.ConstraintTypeUniqueIndex:
			switch len(con.PrimaryColumns) {
			case 0:
			case 1:
				g.Printf(`
unique%s := make(map[%s]struct{})`, formatter.Public(con.PrimaryColumns[0].Name), g.columnType(con.PrimaryColumns[0], pqtgo.ModeDefault))
			default:
			}
		}
	}
	g.Printf(`
	for ent := range model.Generate%sEntity(n) {`, name)
	for _, con := range t.Constraints {
		switch con.Type {
		case pqt.ConstraintTypeUnique, pqt.ConstraintTypeUniqueIndex:
			switch len(con.PrimaryColumns) {
			case 0:
			case 1:
				colName := formatter.Public(con.PrimaryColumns[0].Name)
				g.Printf(`
				if _, ok := unique%s[ent.%s]; ok {
					continue
				} else {
					unique%s[ent.%s] = struct{}{}
				}`, colName, colName, colName, colName)
			default:
			}
		}
	}
	for _, c := range t.Columns {
		colName := formatter.Public(c.Name)
		if c.NotNull {
			switch {
			case g.isNullable(c, pqtgo.ModeDefault):
				g.Printf(`
				if ent.%s == nil {
					continue
				}`, colName)
			case g.isArray(c, pqtgo.ModeDefault):
				g.Printf(`
				if len(ent.%s) == 0 {
					continue
				}`, colName)
			}
		}
	}
	g.Printf(`
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.Get%s().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %%s\nfor entity: %%#v", err.Error(), ent)
			}
			assert%sEntity(t, ent, got)
		})
	}
}`, name, name)
}
