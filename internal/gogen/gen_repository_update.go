package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryUpdateOneByPrimaryKey(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, formatter.Public("updateOneBy", pk.Name), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)
	g.Printf(`
		query, args, err := r.%sQuery(pk, p)
		if err != nil {
			return nil, err
		}`, formatter.Public("updateOneBy", pk.Name))

	g.Printf(`
		var ent %sEntity
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(props...)`,
		entityName,
		formatter.Public("props"),
		formatter.Public("columns"),
		formatter.Public("db"),
	)
	g.Printf(`
		if r.%s != nil {
			r.%s(err, Table%s, "update by primary key", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return &ent, nil
	}`,
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryUpdateOneByPrimaryKeyQuery(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(pk %s, p *%sPatch) (string, []interface{}, error) {`,
		entityName,
		formatter.Public("UpdateOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.Printf(`
		buf := bytes.NewBufferString("UPDATE ")
		buf.WriteString(r.%s)
		update := NewComposer(%d)`,
		formatter.Public("table"),
		len(t.Columns),
	)

	for _, c := range t.Columns {
		g.generateRepositorySetClause(c, "update")
	}
	g.Printf(`
	if !update.Dirty {
		return "", nil, errors.New("%s update failure, nothing to update")
	}`, entityName)

	g.Printf(`
		buf.WriteString(" SET ")
		buf.ReadFrom(update)
		buf.WriteString(" WHERE ")

		update.WriteString(%s)
		update.WriteString("=")
		update.WritePlaceholder()
		update.Add(pk)

		buf.ReadFrom(update)
		buf.WriteString(" RETURNING ")
		if len(r.%s) > 0 {
			buf.WriteString(strings.Join(r.%s, ", "))
		} else {`,
		formatter.Public("table", t.Name, "column", pk.Name),
		formatter.Public("columns"),
		formatter.Public("columns"),
	)

	g.Print(`
		buf.WriteString("`)
	g.selectList(t, -1)
	g.Print(`")
	}`)
	g.Print(`
		return buf.String(), update.Args(), nil
	}`)
}
