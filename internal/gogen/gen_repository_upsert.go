package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
)

func (g *Generator) RepositoryUpsert(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {`,
		entityName,
		formatter.Public("upsert"),
		entityName,
		entityName,
		entityName,
	)
	g.Printf(`
			query, args, err := r.%sQuery(e, p, inf...)
			if err != nil {
				return nil, err
			}
			err = r.%s.QueryRowContext(ctx, query, args...).Scan(`,
		formatter.Public("upsert"),
		formatter.Public("db"),
	)

	for _, c := range t.Columns {
		g.Printf(`
&e.%s,`, formatter.Public(c.Name))
	}
	g.Printf(`
	)
		if r.%s != nil {
			r.%s(err, Table%s, "upsert", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return e, nil
	}`,
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryUpsertQuery(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, p *%sPatch, inf ...string) (string, []interface{}, error) {`,
		entityName,
		formatter.Public("upsert"),
		entityName,
		entityName,
	)
	g.Printf(`
		upsert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns)*2, formatter.Public("table"))

	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositoryInsertClause(c, "upsert")
	}

	g.Print(`
		if upsert.Dirty {
			buf.WriteString(" (")
			buf.ReadFrom(columns)
			buf.WriteString(") VALUES (")
			buf.ReadFrom(upsert)
			buf.WriteString(")")
		}
		buf.WriteString(" ON CONFLICT ")`,
	)

	g.Print(`
		if len(inf) > 0 {
		upsert.Dirty=false`)
	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositorySetClause(c, "upsert")
	}
	closeBrace(g, 1)

	g.Printf(`
		if len(inf) > 0 && upsert.Dirty {
			buf.WriteString("(")
			for j, i := range inf {
				if j != 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(i)
			}
			buf.WriteString(")")
			buf.WriteString(" DO UPDATE SET ")
			buf.ReadFrom(upsert)
		} else {
			buf.WriteString(" DO NOTHING ")
		}
		if upsert.Dirty {
			buf.WriteString(" RETURNING ")
			if len(r.%s) > 0 {
				buf.WriteString(strings.Join(r.%s, ", "))
			} else {`,
		formatter.Public("columns"),
		formatter.Public("columns"),
	)
	g.Print(`
		buf.WriteString("`)
	g.selectList(t, -1)
	g.Print(`")
	}`)
	g.Print(`
		}
		return buf.String(), upsert.Args(), nil
	}`)
}
