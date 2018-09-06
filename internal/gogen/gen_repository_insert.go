package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) RepositoryMethodInsert(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity) (*%sEntity, error) {`, entityName, pqtfmt.Public("insert"), entityName, entityName)
	g.Printf(`
			return r.%s(ctx, nil, e)
		}`,
		pqtfmt.Private("insert"),
	)
}

func (g *Generator) RepositoryTxMethodInsert(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBaseTx) %s(ctx context.Context, e *%sEntity) (*%sEntity, error) {`, entityName, pqtfmt.Public("insert"), entityName, entityName)
	g.Printf(`
			return r.base.%s(ctx, r.tx, e)
		}`,
		pqtfmt.Private("insert"),
	)
}

func (g *Generator) RepositoryMethodPrivateInsert(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, e *%sEntity) (*%sEntity, error) {`, entityName, pqtfmt.Private("insert"), entityName, entityName)
	g.Printf(`
			query, args, err := r.%sQuery(e, true)
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
		pqtfmt.Public("insert"),
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
				r.%s(err, Table%s, "insert", query, args...)
			} else {
				r.%s(err, Table%s, "insert tx", query, args...)
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

func (g *Generator) RepositoryMethodInsertQuery(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, read bool) (string, []interface{}, error) {`, entityName, pqtfmt.Public("insert"), entityName)
	g.Printf(`
		insert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns), pqtfmt.Public("table"))

	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositoryInsertClause(c, "insert")
	}
	g.Print(`
		if columns.Len() > 0 {
			buf.WriteString(" (")
			buf.ReadFrom(columns)
			buf.WriteString(") VALUES (")
			buf.ReadFrom(insert)
			buf.WriteString(") ")`)
	g.Printf(`
			if read {
				buf.WriteString("RETURNING ")
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
		}
		return buf.String(), insert.Args(), nil
	}`)
}
