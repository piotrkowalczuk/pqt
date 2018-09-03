package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) RepositoryMethodUpsert(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {
			return r.%s(ctx, nil, e, p, inf...)
		}`,
		entityName,
		pqtfmt.Public("upsert"),
		entityName,
		entityName,
		entityName,
		pqtfmt.Private("upsert"),
	)
}

func (g *Generator) RepositoryTxMethodUpsert(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBaseTx) %s(ctx context.Context, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {
			return r.base.%s(ctx, r.tx, e, p, inf...)
		}`,
		entityName,
		pqtfmt.Public("upsert"),
		entityName,
		entityName,
		entityName,
		pqtfmt.Private("upsert"),
	)
}

func (g *Generator) RepositoryMethodPrivateUpsert(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {`,
		entityName,
		pqtfmt.Private("upsert"),
		entityName,
		entityName,
		entityName,
	)
	g.Printf(`
			query, args, err := r.%sQuery(e, p, inf...)
			if err != nil {
				return nil, err
			}

			var row *sql.Row
			if tx == nil {
				row = r.%s.QueryRowContext(ctx, query, args...)
			} else {
				row = tx.QueryRowContext(ctx, query, args...)
			}
			err = row.Scan(`,
		pqtfmt.Public("upsert"),
		pqtfmt.Public("db"),
	)

	for _, c := range t.Columns {
		g.Printf(`
&e.%s,`, pqtfmt.Public(c.Name))
	}
	g.Printf(`
	)
		if r.%s != nil {
			if tx == nil {
				r.%s(err, Table%s, "upsert", query, args...)
			} else {
				r.%s(err, Table%s, "upsert tx", query, args...)
			}
		}
		if err != nil {
			return nil, err
		}
		return e, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
		pqtfmt.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryMethodUpsertQuery(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, p *%sPatch, inf ...string) (string, []interface{}, error) {`,
		entityName,
		pqtfmt.Public("upsert"),
		entityName,
		entityName,
	)
	g.Printf(`
		upsert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns)*2, pqtfmt.Public("table"))

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
		pqtfmt.Public("columns"),
		pqtfmt.Public("columns"),
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
