package gogen

import (
	"fmt"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryMethodUpdateOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, pqtfmt.Public("updateOneBy", pk.Name), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)
	g.Printf(`
		return r.%s(ctx, nil, pk, p)
		}`,
		pqtfmt.Private("updateOneBy", pk.Name),
	)
}

func (g *Generator) RepositoryTxMethodUpdateOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBaseTx) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, pqtfmt.Public("updateOneBy", pk.Name), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)
	g.Printf(`
		return r.base.%s(ctx, r.tx, pk, p)
		}`,
		pqtfmt.Private("updateOneBy", pk.Name),
	)
}

func (g *Generator) RepositoryMethodPrivateUpdateOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, pqtfmt.Private("updateOneBy", pk.Name), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)
	g.Printf(`
		query, args, err := r.%sQuery(pk, p)
		if err != nil {
			return nil, err
		}`, pqtfmt.Public("updateOneBy", pk.Name))

	g.Printf(`
		var ent %sEntity
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}`,
		entityName,
		pqtfmt.Public("props"),
		pqtfmt.Public("columns"),
	)
	g.Printf(`
		if tx == nil {
			err = r.%s.QueryRowContext(ctx, query, args...).Scan(props...)
		} else {
			err = tx.QueryRowContext(ctx, query, args...).Scan(props...)
		}`,
		pqtfmt.Public("db"))
	g.Printf(`
		if r.%s != nil {
			if tx == nil {
				r.%s(err, Table%s, "update by primary key", query, args...)
			} else {
				r.%s(err, Table%s, "update by primary key tx", query, args...)
			}
		}
		if err != nil {
			return nil, err
		}
		return &ent, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
		pqtfmt.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryMethodUpdateOneByPrimaryKeyQuery(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(pk %s, p *%sPatch) (string, []interface{}, error) {`,
		entityName,
		pqtfmt.Public("UpdateOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.Printf(`
		buf := bytes.NewBufferString("UPDATE ")
		buf.WriteString(r.%s)
		update := NewComposer(%d)`,
		pqtfmt.Public("table"),
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
		pqtfmt.Public("table", t.Name, "column", pk.Name),
		pqtfmt.Public("columns"),
		pqtfmt.Public("columns"),
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

func (g *Generator) RepositoryMethodUpdateOneByUniqueConstraintQuery(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

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
			arguments += fmt.Sprintf("%s %s", pqtfmt.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		method = append(method, "query")

		g.Printf(`
			func (r *%sRepositoryBase) %s(%s, p *%sPatch) (string, []interface{}, error) {`,
			entityName,
			pqtfmt.Public(method...),
			arguments,
			entityName,
		)

		g.Printf(`
			buf := bytes.NewBufferString("UPDATE ")
			buf.WriteString(r.%s)
			update := NewComposer(%d)`, pqtfmt.Public("table"), len(u.PrimaryColumns))

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
				pqtfmt.Public("table", t.Name, "column", c.Name),
				pqtfmt.Private(columnForeignName(c)),
			)
		}
		g.Printf(`
			buf.ReadFrom(update)
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

func (g *Generator) RepositoryMethodUpdateOneByUniqueConstraint(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	for i, u := range uniqueConstraints(t) {
		if i > 0 {
			g.NewLine()
		}
		method := []string{"updateOneBy"}
		arguments := ""
		argumentsNameOnly := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
				argumentsNameOnly += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", pqtfmt.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
			argumentsNameOnly += pqtfmt.Private(columnForeignName(c))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}
		g.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, %s, p *%sPatch) (*%sEntity, error) {
				return r.%s(ctx, nil, %s, p)
			}`,
			entityName,
			pqtfmt.Public(method...),
			arguments,
			entityName,
			entityName,
			pqtfmt.Private(method...),
			argumentsNameOnly,
		)
	}
}

func (g *Generator) RepositoryTxMethodUpdateOneByUniqueConstraint(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	for i, u := range uniqueConstraints(t) {
		if i > 0 {
			g.NewLine()
		}
		method := []string{"updateOneBy"}
		arguments := ""
		argumentsNameOnly := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
				argumentsNameOnly += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", pqtfmt.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
			argumentsNameOnly += pqtfmt.Private(columnForeignName(c))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}
		g.Printf(`
			func (r *%sRepositoryBaseTx) %s(ctx context.Context, %s, p *%sPatch) (*%sEntity, error) {
				return r.base.%s(ctx, r.tx, %s, p)
			}`,
			entityName,
			pqtfmt.Public(method...),
			arguments,
			entityName,
			entityName,
			pqtfmt.Private(method...),
			argumentsNameOnly,
		)
	}
}

func (g *Generator) RepositoryMethodPrivateUpdateOneByUniqueConstraint(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
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
			arguments += fmt.Sprintf("%s %s", pqtfmt.Private(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
			arguments2 += pqtfmt.Private(columnForeignName(c))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		g.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, %s, p *%sPatch) (*%sEntity, error) {`,
			entityName,
			pqtfmt.Private(method...),
			arguments,
			entityName,
			entityName,
		)

		g.Printf(`
			query, args, err := r.%s(%s, p)
			if err != nil {
				return nil, err
			}`,
			pqtfmt.Public(append(method, "query")...),
			arguments2,
		)
		g.Printf(`
			var ent %sEntity
			props, err := ent.%s(r.%s...)
			if err != nil {
				return nil, err
			}

			var row *sql.Row
			if tx == nil {
				row = r.%s.QueryRowContext(ctx, query, args...)
			} else {
				row = tx.QueryRowContext(ctx, query, args...)
			}`,
			entityName,
			pqtfmt.Public("props"),
			pqtfmt.Public("columns"),
			pqtfmt.Public("db"),
		)

		g.Printf(`
				err = row.Scan(props...)
				if r.%s != nil {
					if tx == nil {
						r.%s(err, Table%s, "update one by unique", query, args...)
					} else {
						r.%s(err, Table%s, "update one by unique tx", query, args...)
					}
				}
				if err != nil {
					return nil, err
				}
				return &ent, nil
			}`,
			pqtfmt.Public("log"),
			pqtfmt.Public("log"),
			entityName,
			pqtfmt.Public("log"),
			entityName,
		)
	}
}
