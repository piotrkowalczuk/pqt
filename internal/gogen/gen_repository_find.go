package gogen

import (
	"fmt"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryFindIter(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) (*%sIterator, error) {`, entityName, formatter.Public("findIter"), entityName, entityName)
	g.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		formatter.Public("find"),
		formatter.Public("db"),
	)

	g.Printf(`
	 	if r.%s != nil {
			r.%s(err, Table%s, "find iter", query, args...)
		}
		if err != nil {
			return nil, err
		}`,
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
	g.Printf(`
			return &%sIterator{
				rows: rows,
				expr: fe,
				cols: fe.Columns,
		}, nil
	}`, formatter.Public(t.Name))
}

func (g *Generator) RepositoryFindQuery(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(fe *%sFindExpr) (string, []interface{}, error) {`, entityName, formatter.Public("find"), entityName)
	g.Printf(`
		comp := NewComposer(%d)
		buf := bytes.NewBufferString("SELECT ")
		if len(fe.%s) == 0 {
		buf.WriteString("`, len(t.Columns), formatter.Public("columns"))
	g.selectList(t, 0)
	g.Printf(`")
		} else {
			buf.WriteString(strings.Join(fe.%s, ", "))
		}`, formatter.Public("columns"))
	for nb, r := range joinableRelationships(t) {
		g.Printf(`
			if fe.%s != nil && fe.%s.%s {`,
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("fetch"),
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
		buf.WriteString(" AS t0")`, formatter.Public("table"))
	for nb, r := range joinableRelationships(t) {
		oc := r.OwnerColumns
		ic := r.InversedColumns
		if len(oc) != len(ic) {
			panic("number of owned and inversed foreign key columns is not equal")
		}

		g.Printf(`
			if fe.%s != nil {`,
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
		)
		g.Printf(`
			joinClause(comp, fe.%s.%s, "%s AS t%d ON `,
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("kind"),
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
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("on"),
			formatter.Public(r.InversedTable.Name),
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("on"),
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
		formatter.Public("where"),
		formatter.Public(t.Name),
		formatter.Public("where"),
	)

	for nb, r := range joinableRelationships(t) {
		g.Printf(`
		if fe.%s != nil && fe.%s.%s != nil {
			if err := %sCriteriaWhereClause(comp, fe.%s.%s, %d); err != nil {
				return "", nil, err
			}
		}`,
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("where"),
			formatter.Public(r.InversedTable.Name),
			formatter.Public("join", or(r.InversedName, r.InversedTable.Name)),
			formatter.Public("where"),
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
		formatter.Public("orderBy"),
		formatter.Public("orderBy"),
		formatter.Public("table", t.Name, "columns"),
		formatter.Public("offset"),
		formatter.Public("offset"),
		formatter.Public("limit"),
		formatter.Public("limit"),
	)

	g.Print(`
	buf.ReadFrom(comp)

	return buf.String(), comp.Args(), nil
}`)
}

func (g *Generator) RepositoryFindOneByPrimaryKey(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (*%sEntity, error) {`,
		entityName,
		formatter.Public("FindOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.Printf(`
		find := NewComposer(%d)
		find.WriteString("SELECT ")
		if len(r.%s) == 0 {
			find.WriteString("`,
		len(t.Columns), formatter.Public("columns"))
	g.selectList(t, -1)
	g.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, formatter.Public("columns"))

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
		formatter.Public("table", t.Name),
		formatter.Public("table", t.Name, "column", pk.Name),
		entityName,
	)

	g.Printf(`
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
		formatter.Public("props"),
		formatter.Public("columns"),
		formatter.Public("db"),
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
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryFind(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) ([]*%sEntity, error) {`, entityName, formatter.Public("find"), entityName, entityName)
	g.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		formatter.Public("find"),
		formatter.Public("db"),
	)

	g.Printf(`
		if r.%s != nil {
			r.%s(err, Table%s, "find", query, args...)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()`,
		formatter.Public("log"),
		formatter.Public("log"),
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
		formatter.Public(t.Name),
		formatter.Public("props"),
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
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
}

func (g *Generator) RepositoryFindOneByUniqueConstraint(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	for _, u := range uniqueConstraints(t) {
		method := []string{"FindOneBy"}
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

		g.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, %s) (*%sEntity, error) {`,
			entityName,
			formatter.Public(method...),
			arguments,
			entityName,
		)
		g.Printf(`
			find := NewComposer(%d)
			find.WriteString("SELECT ")
					if len(r.%s) == 0 {
			find.WriteString("`,
			len(t.Columns), formatter.Public("columns"))
		g.selectList(t, -1)
		g.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, formatter.Public("columns"))

		partialClause := ""
		if len(u.Where) > 0 {
			partialClause = fmt.Sprintf("%s AND ", u.Where)
		}

		g.Printf(`
			find.WriteString(" FROM ")
			find.WriteString(%s)
			find.WriteString(" WHERE %s")`,
			formatter.Public("table", t.Name),
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
		`, formatter.Public("table", t.Name, "column", c.Name), formatter.Private(columnForeignName(c)))
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
			formatter.Public("props"),
			formatter.Public("columns"),
			formatter.Public("db"),
		)
		g.Print(`
			if err != nil {
				return nil, err
			}

			return &ent, nil
		}`)
	}
}
