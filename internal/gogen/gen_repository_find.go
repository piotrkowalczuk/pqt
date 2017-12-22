package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
)

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
