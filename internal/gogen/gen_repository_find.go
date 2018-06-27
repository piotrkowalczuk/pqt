package gogen

import (
	"fmt"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryFindIter(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) (*%sIterator, error) {`, entityName, pqtfmt.Public("findIter"), entityName, entityName)
	g.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		pqtfmt.Public("find"),
		pqtfmt.Public("db"),
	)

	g.Printf(`
	 	if r.%s != nil {
			r.%s(err, Table%s, "find iter", query, args...)
		}
		if err != nil {
			return nil, err
		}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)
	g.Printf(`
			return &%sIterator{
				rows: rows,
				expr: fe,
				cols: fe.Columns,
		}, nil
	}`, pqtfmt.Public(t.Name))
}

func (g *Generator) RepositoryFindQuery(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(fe *%sFindExpr) (string, []interface{}, error) {`, entityName, pqtfmt.Public("find"), entityName)
	g.Printf(`
		comp := NewComposer(%d)
		buf := bytes.NewBufferString("SELECT ")
		if len(fe.%s) == 0 {
		buf.WriteString("`, len(t.Columns), pqtfmt.Public("columns"))
	g.selectList(t, 0)
	g.Printf(`")
		} else {
			buf.WriteString(strings.Join(fe.%s, ", "))
		}`, pqtfmt.Public("columns"))
	// Generate select clause for joinable tables if needed.
	for nb, r := range joinableRelationships(t) {
		joinPropertyName := pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name))

		g.Printf(`
			if fe.%s != nil && fe.%s.Kind.Actionable() && fe.%s.%s {`,
			joinPropertyName,
			joinPropertyName,
			joinPropertyName,
			pqtfmt.Public("fetch"),
		)
		g.Print(`
		buf.WriteString(", `)
		g.selectList(r.InversedTable, nb+1)
		g.Print(`")`)
		closeBrace(g, 1)
	}
	g.Printf(`
		buf.WriteString(" FROM ")
		buf.WriteString(r.%s)
		buf.WriteString(" AS t0")`, pqtfmt.Public("table"))
	// Generate JOIN clause for joinable tables if needed.
	for nb, r := range joinableRelationships(t) {
		oc := r.OwnerColumns
		ic := r.InversedColumns
		joinPropertyName := pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name))

		if len(oc) != len(ic) {
			panic("number of owned and inversed foreign key columns is not equal")
		}

		g.Printf(`
			if fe.%s != nil && fe.%s.Kind.Actionable() {`,
			joinPropertyName,
			joinPropertyName,
		)
		g.Printf(`
			joinClause(comp, fe.%s.%s, "%s AS t%d ON `,
			joinPropertyName,
			pqtfmt.Public("kind"),
			r.InversedTable.FullName(),
			nb+1,
		)

		for i := 0; i < len(oc); i++ {
			if i > 0 {
				g.Print(` AND `)
			}
			g.Printf(`t%d.%s=t%d.%s`, 0, oc[i].Name, nb+1, ic[i].Name)
		}
		g.Print(`")`)

		g.Printf(`
		if fe.%s.%s != nil {
			comp.Dirty = true
			if err := %sCriteriaWhereClause(comp, fe.%s.%s, %d); err != nil {
				return "", nil, err
			}
		}`,
			joinPropertyName,
			pqtfmt.Public("on"),
			pqtfmt.Public(r.InversedTable.Name),
			joinPropertyName,
			pqtfmt.Public("on"),
			nb+1,
		)

		closeBrace(g, 1)
	}

	g.Printf(`
	if comp.Dirty {
		buf.ReadFrom(comp)
		comp.Dirty = false
	}
	if fe.%s != nil {
		if err := %sCriteriaWhereClause(comp, fe.%s, 0); err != nil {
			return "", nil, err
		}
	}`,
		pqtfmt.Public("where"),
		pqtfmt.Public(t.Name),
		pqtfmt.Public("where"),
	)

	for nb, r := range joinableRelationships(t) {
		joinPropertyName := pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name))
		g.Printf(`
		if fe.%s != nil && fe.%s.Kind.Actionable() && fe.%s.%s != nil {
			if err := %sCriteriaWhereClause(comp, fe.%s.%s, %d); err != nil {
				return "", nil, err
			}
		}`,
			joinPropertyName,
			joinPropertyName,
			joinPropertyName,
			pqtfmt.Public("where"),
			pqtfmt.Public(r.InversedTable.Name),
			joinPropertyName,
			pqtfmt.Public("where"),
			nb+1,
		)
	}

	g.Print(`
		if comp.Dirty {
			if _, err := buf.WriteString(" WHERE "); err != nil {
				return "", nil, err
			}
			buf.ReadFrom(comp)
		}
	`)

	g.Printf(`
	if len(fe.%s) > 0 {
		i:=0
		for _, order := range fe.%s {
			for _, columnName := range %s {
				if order.Name == columnName {
					if i == 0 {
						comp.WriteString(" ORDER BY ")
					}
					if i > 0 {
						if _, err := comp.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if _, err := comp.WriteString(order.Name); err != nil {
						return "", nil, err
					}
					if order.Descending {
						if _, err := comp.WriteString(" DESC"); err != nil {
							return "", nil, err
						}
					}
					i++
					break
				}
			}
		}
	}
	if fe.%s > 0 {
		if _, err := comp.WriteString(" OFFSET "); err != nil {
			return "", nil, err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		if _, err := comp.WriteString(" "); err != nil {
			return "", nil, err
		}
		comp.Add(fe.%s)
	}
	if fe.%s > 0 {
		if _, err := comp.WriteString(" LIMIT "); err != nil {
			return "", nil, err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		if _, err := comp.WriteString(" "); err != nil {
			return "", nil, err
		}
		comp.Add(fe.%s)
	}
`,
		pqtfmt.Public("orderBy"),
		pqtfmt.Public("orderBy"),
		pqtfmt.Public("table", t.Name, "columns"),
		pqtfmt.Public("offset"),
		pqtfmt.Public("offset"),
		pqtfmt.Public("limit"),
		pqtfmt.Public("limit"),
	)

	g.Print(`
	buf.ReadFrom(comp)

	return buf.String(), comp.Args(), nil
}`)
}

func (g *Generator) RepositoryFindOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (*%sEntity, error) {`,
		entityName,
		pqtfmt.Public("FindOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.Printf(`
		find := NewComposer(%d)
		find.WriteString("SELECT ")
		if len(r.%s) == 0 {
			find.WriteString("`,
		len(t.Columns), pqtfmt.Public("columns"))
	g.selectList(t, -1)
	g.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, pqtfmt.Public("columns"))

	g.Printf(`
		find.WriteString(" FROM ")
		find.WriteString(%s)
		find.WriteString(" WHERE ")
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(pk)
		var (
			ent %sEntity
		)`,
		pqtfmt.Public("table", t.Name),
		pqtfmt.Public("table", t.Name, "column", pk.Name),
		entityName,
	)

	g.Printf(`
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
		pqtfmt.Public("props"),
		pqtfmt.Public("columns"),
		pqtfmt.Public("db"),
	)
	g.Printf(`
		if r.%s != nil {
			r.%s(err, Table%s, "find by primary key", find.String(), find.Args()...)
		}
		if err != nil {
			return nil, err
		}
		return &ent, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryFindOneByPrimaryKeyAndUpdate(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (before, after *%sEntity, err error) {`, entityName, pqtfmt.Public("findOneBy", pk.Name, "AndUpdate"), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)

	g.Printf(`
		find := NewComposer(%d)
		find.WriteString("SELECT ")
		if len(r.%s) == 0 {
			find.WriteString("`,
		len(t.Columns), pqtfmt.Public("columns"))
	g.selectList(t, -1)
	g.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, pqtfmt.Public("columns"))

	g.Printf(`
		find.WriteString(" FROM ")
		find.WriteString(%s)
		find.WriteString(" WHERE ")
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(pk)
		find.WriteString(" FOR UPDATE")`,
		pqtfmt.Public("table", t.Name),
		pqtfmt.Public("table", t.Name, "column", pk.Name),
	)
	g.Printf(`
		query, args, err := r.%sQuery(pk, p)
		if err != nil {
			return
		}`, pqtfmt.Public("updateOneBy", pk.Name))

	g.Printf(`
		var (
			oldEnt, newEnt %sEntity
		)
		oldProps, err := oldEnt.%s(r.%s...)
		if err != nil {
			return
		}
		newProps, err := newEnt.%s(r.%s...)
		if err != nil {
			return
		}`,
		entityName,
		pqtfmt.Public("props"),
		pqtfmt.Public("columns"),
		pqtfmt.Public("props"),
		pqtfmt.Public("columns"),
	)

	g.Printf(`
		tx, err := r.%s.Begin()
		if err != nil {
			return
		}`,
		pqtfmt.Public("db"),
	)

	g.Printf(`
		err = tx.QueryRowContext(ctx, find.String(), find.Args()...).Scan(oldProps...)
		if r.%s != nil {
			r.%s(err, Table%s, "find by primary key", find.String(), find.Args()...)
		}
		if err != nil {
			tx.Rollback()
			return
		}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)
	g.Printf(`
		err = tx.QueryRowContext(ctx, query, args...).Scan(newProps...)
		if r.%s != nil {
			r.%s(err, Table%s, "update by primary key", query, args...)
		}
		if err != nil {
			tx.Rollback()
			return
		}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)

	g.Printf(`
		err = tx.Commit()
		if err != nil {
			return
		}
		return &oldEnt, &newEnt, nil
	}`)
}

func (g *Generator) RepositoryFind(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) ([]*%sEntity, error) {`, entityName, pqtfmt.Public("find"), entityName, entityName)
	g.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		pqtfmt.Public("find"),
		pqtfmt.Public("db"),
	)

	g.Printf(`
		if r.%s != nil {
			r.%s(err, Table%s, "find", query, args...)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)

	g.Printf(`
		var (
			entities []*%sEntity
			props []interface{}
		)
		for rows.Next() {
			var ent %sEntity
			if props, err = ent.%s(); err != nil {
				return nil, err
			}`,
		entityName,
		pqtfmt.Public(t.Name),
		pqtfmt.Public("props"),
	)
	if hasJoinableRelationships(t) {
		g.Print(`
		var prop []interface{}`)
	}
	g.scanJoinableRelationships(t, "fe")
	g.Print(`
			err = rows.Scan(props...)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &ent)
		}`)
	g.Printf(`
		err = rows.Err()
		if r.%s != nil {
			r.%s(err, Table%s, "find", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return entities, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryFindOneByUniqueConstraint(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	for _, u := range uniqueConstraints(t) {
		method := []string{"FindOneBy"}
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

		g.Printf(`

			func (r *%sRepositoryBase) %s(ctx context.Context, %s) (*%sEntity, error) {`,
			entityName,
			pqtfmt.Public(method...),
			arguments,
			entityName,
		)
		g.Printf(`
			find := NewComposer(%d)
			find.WriteString("SELECT ")
					if len(r.%s) == 0 {
			find.WriteString("`,
			len(t.Columns), pqtfmt.Public("columns"))
		g.selectList(t, -1)
		g.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, pqtfmt.Public("columns"))

		partialClause := ""
		if len(u.Where) > 0 {
			partialClause = fmt.Sprintf("%s AND ", u.Where)
		}

		g.Printf(`
			find.WriteString(" FROM ")
			find.WriteString(%s)
			find.WriteString(" WHERE %s")`,
			pqtfmt.Public("table", t.Name),
			partialClause,
		)
		for i, c := range u.PrimaryColumns {
			if i != 0 {
				g.Print(`find.WriteString(" AND ")`)
			}
			g.Printf(`
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(%s)
		`, pqtfmt.Public("table", t.Name, "column", c.Name), pqtfmt.Private(columnForeignName(c)))
		}

		g.Printf(`
			var (
				ent %sEntity
			)
			props, err := ent.%s(r.%s...)
			if err != nil {
				return nil, err
			}
			err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
			entityName,
			pqtfmt.Public("props"),
			pqtfmt.Public("columns"),
			pqtfmt.Public("db"),
		)
		g.Print(`
			if err != nil {
				return nil, err
			}

			return &ent, nil
		}`)
	}
}
