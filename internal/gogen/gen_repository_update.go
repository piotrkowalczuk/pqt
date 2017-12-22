package gogen

import (
	"fmt"

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

func (g *Generator) RepositoryUpdateOneByUniqueConstraintQuery(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	for i, u := range uniqueConstraints(t) {
		if i > 0 {
			g.NewLine()
		}
		method := []string{"updateOneBy"}
		arguments := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", formatter.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		method = append(method, "query")

		g.Printf(`
			func (r *%sRepositoryBase) %s(%s, p *%sPatch) (string, []interface{}, error) {`,
			entityName,
			formatter.Public(method...),
			arguments,
			entityName,
		)

		g.Printf(`
			buf := bytes.NewBufferString("UPDATE ")
			buf.WriteString(r.%s)
			update := NewComposer(%d)`, formatter.Public("table"), len(u.PrimaryColumns))

		for _, c := range t.Columns {
			g.generateRepositorySetClause(c, "update")
		}
		g.Printf(`
			if !update.Dirty {
				return "", nil, errors.New("%s update failure, nothing to update")
			}`, t.Name,
		)
		g.Print(`
			buf.WriteString(" SET ")
			buf.ReadFrom(update)
			buf.WriteString(" WHERE ")`)
		for i, c := range u.PrimaryColumns {
			if i != 0 {
				g.Print(`
					update.WriteString(" AND ")`)
			}
			g.Printf(`
				update.WriteString(%s)
				update.WriteString("=")
				update.WritePlaceholder()
				update.Add(%s)`,
				formatter.Public("table", t.Name, "column", c.Name),
				formatter.Private(columnForeignName(c)),
			)
		}
		g.Printf(`
			buf.ReadFrom(update)
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
		if len(u.Where) > 0 {
			g.Printf(` WHERE %s")
	}`, u.Where)
		} else {
			g.Print(`")
	}`)
		}
		g.Print(`
		return buf.String(), update.Args(), nil
	}`)
	}
}

func (g *Generator) RepositoryUpdateOneByUniqueConstraint(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	for i, u := range uniqueConstraints(t) {
		if i > 0 {
			g.NewLine()
		}
		method := []string{"updateOneBy"}
		arguments := ""
		arguments2 := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
				arguments2 += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", formatter.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
			arguments2 += formatter.Private(columnForeignName(c))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		g.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, %s, p *%sPatch) (*%sEntity, error) {`,
			entityName,
			formatter.Public(method...),
			arguments,
			entityName,
			entityName,
		)

		g.Printf(`
			query, args, err := r.%s(%s, p)
			if err != nil {
				return nil, err
			}`,
			formatter.Public(append(method, "query")...),
			arguments2,
		)
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
				r.%s(err, Table%s, "update one by unique", query, args...)
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
}
