package gogen

import (
	"strings"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryInsertQuery(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, read bool) (string, []interface{}, error) {`, entityName, formatter.Public("insert"), entityName)
	g.Printf(`
		insert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns), formatter.Public("table"))

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
		}
		return buf.String(), insert.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryInsertClause(c *pqt.Column, sel string) {
	braces := 0

	switch c.Type {
	case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
		return
	default:
		if g.canBeNil(c, pqtgo.ModeDefault) {
			g.Printf(`
					if e.%s != nil {`,
				formatter.Public(c.Name),
			)
			braces++
		}
		if g.isNullable(c, pqtgo.ModeDefault) {
			g.Printf(`
					if e.%s.Valid {`, formatter.Public(c.Name))
			braces++
		}
		if g.isType(c, pqtgo.ModeDefault, "time.Time") {
			g.Printf(`
					if !e.%s.IsZero() {`, formatter.Public(c.Name))
			braces++
		}
		g.Printf(strings.Replace(`
			if columns.Len() > 0 {
				if _, err := columns.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := columns.WriteString(%s); err != nil {
				return "", nil, err
			}
			if {{SELECTOR}}.Dirty {
				if _, err := {{SELECTOR}}.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if err := {{SELECTOR}}.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			{{SELECTOR}}.Add(e.%s)
			{{SELECTOR}}.Dirty=true`, "{{SELECTOR}}", sel, -1),
			formatter.Public("table", c.Table.Name, "column", c.Name),
			formatter.Public(c.Name),
		)

		closeBrace(g, braces)
		g.Println("")
	}
}
