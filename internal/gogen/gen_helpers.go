package gogen

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RunInTransaction() {
	g.Printf(`
func RunInTransaction(db *sql.DB, ctx context.Context, f func(tx *sql.Tx) error, attempts int) (err error) {
	for n := 0; n < attempts; n++ {
		if err = func () error {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}

			defer func() {
				if p := recover(); p != nil {
					_ = tx.Rollback()
					panic(p)
				} else if err != nil {
					_ = tx.Rollback()
				}
			}()

			if err = f(tx); err != nil {
				_ = tx.Rollback()
				return err
			}

			return tx.Commit()
		}(); err == RetryTransaction {
			continue
		}
		return err
	}
	return err
}`)
}

func (g *Generator) isArray(c *pqt.Column, m int32) bool {
	if strings.HasPrefix(g.columnType(c, m), "[]") {
		return true
	}

	return g.isType(c, m, "pq.StringArray", "pq.Int64Array", "pq.BoolArray", "pq.Float64Array", "pq.ByteaArray", "pq.GenericArray")
}

func (g *Generator) columnType(c *pqt.Column, m int32) string {
	m = columnMode(c, m)
	for _, plugin := range g.Plugins {
		if txt := plugin.PropertyType(c, m); txt != "" {
			return txt
		}
	}
	res := pqtfmt.Type(c.Type, m)
	if res == "" {
		res = "<nil>"
	}
	return res
}

func (g *Generator) isType(c *pqt.Column, m int32, types ...string) bool {
	for _, t := range types {
		if g.columnType(c, m) == t {
			return true
		}
	}
	return false
}

func (g *Generator) isNullable(c *pqt.Column, m int32) bool {
	if mt, ok := c.Type.(pqt.MappableType); ok {
		for _, mapto := range mt.Mapping {
			if ct, ok := mapto.(pqtgo.CustomType); ok {
				tof := ct.TypeOf(columnMode(c, m))
				if tof == nil {
					continue
				}
				if tof.Kind() != reflect.Struct {
					continue
				}

				if field, ok := tof.FieldByName("Valid"); ok {
					if field.Type.Kind() == reflect.Bool {
						return true
					}
				}
			}
		}
	}
	return g.isType(c, m,
		// sql
		"sql.NullString",
		"sql.NullBool",
		"sql.NullInt64",
		"sql.NullFloat64",
		// pq
		"pq.NullTime",
		// generated
		"NullInt64Array",
		"NullFloat64Array",
		"NullBoolArray",
		"NullByteaArray",
		"NullStringArray",
		"NullBoolArray",
	)
}

func (g *Generator) canBeNil(c *pqt.Column, m int32) bool {
	if tp, ok := c.Type.(pqt.MappableType); ok {
		for _, mapto := range tp.Mapping {
			if ct, ok := mapto.(pqtgo.CustomType); ok {
				switch m {
				case pqtgo.ModeMandatory:
					return ct.TypeOf(pqtgo.ModeMandatory).Kind() == reflect.Ptr || ct.TypeOf(pqtgo.ModeMandatory).Kind() == reflect.Map
				case pqtgo.ModeOptional:
					return ct.TypeOf(pqtgo.ModeOptional).Kind() == reflect.Ptr || ct.TypeOf(pqtgo.ModeOptional).Kind() == reflect.Map
				case pqtgo.ModeCriteria:
					return ct.TypeOf(pqtgo.ModeCriteria).Kind() == reflect.Ptr || ct.TypeOf(pqtgo.ModeCriteria).Kind() == reflect.Map
				default:
					return false
				}
			}
		}
	}
	if g.columnType(c, m) == "interface{}" {
		return true
	}
	if strings.HasPrefix(g.columnType(c, m), "*") {
		return true
	}
	if g.isArray(c, m) {
		return true
	}
	if g.isType(c, m,
		"pq.StringArray",
		"ByteaArray",
		"pq.BoolArray",
		"pq.Int64Array",
		"pq.Float64Array",
	) {
		return true
	}
	return false
}

func (g *Generator) selectList(t *pqt.Table, nb int) {
	for i, c := range t.Columns {
		if i != 0 {
			g.Print(", ")
		}
		if c.IsDynamic {
			g.Printf("%s(", c.Func.Name)
			for i, arg := range c.Func.Args {
				if arg.Type != c.Columns[i].Type {
					fmt.Printf("wrong function (%s) argument type, expected %v but got %v\n", c.Func.Name, arg.Type, c.Columns[i].Type)
				}
				if i != 0 {
					g.Print(", ")
				}
				if nb > -1 {
					g.Printf("t%d.%s", nb, c.Columns[i].Name)
				} else {
					g.Printf("%s", c.Columns[i].Name)
				}
			}
			g.Printf(") AS %s", c.Name)
		} else {
			if nb > -1 {
				g.Printf("t%d.%s", nb, c.Name)
			} else {
				g.Printf("%s", c.Name)
			}
		}
	}
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
				pqtfmt.Public(c.Name),
			)
			braces++
		}
		if g.isNullable(c, pqtgo.ModeDefault) {
			g.Printf(`
					if e.%s.Valid {`, pqtfmt.Public(c.Name))
			braces++
		}
		if g.isType(c, pqtgo.ModeDefault, "time.Time") {
			g.Printf(`
					if !e.%s.IsZero() {`, pqtfmt.Public(c.Name))
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
			pqtfmt.Public("table", c.Table.Name, "column", c.Name),
			pqtfmt.Public(c.Name),
		)

		closeBrace(g, braces)
		g.NewLine()
	}
}

func (g *Generator) generateRepositorySetClause(c *pqt.Column, sel string) {
	if c.PrimaryKey {
		return
	}
	for _, plugin := range g.Plugins {
		if txt := plugin.SetClause(c); txt != "" {
			tmpl, err := template.New("root").Parse(txt)
			if err != nil {
				panic(err)
			}
			if err = tmpl.Execute(g, map[string]interface{}{
				"selector": fmt.Sprintf("p.%s", pqtfmt.Public(c.Name)),
				"column":   pqtfmt.Public("table", c.Table.Name, "column", c.Name),
				"composer": sel,
			}); err != nil {
				panic(err)
			}
			g.Println("")
			return
		}
	}
	braces := 0
	if g.canBeNil(c, pqtgo.ModeOptional) {
		g.Printf(`
			if p.%s != nil {`, pqtfmt.Public(c.Name))
		braces++
	}
	if g.isNullable(c, pqtgo.ModeOptional) {
		g.Printf(`
			if p.%s.Valid {`, pqtfmt.Public(c.Name))
		braces++
	}
	if g.isType(c, pqtgo.ModeOptional, "time.Time") {
		g.Printf(`
			if !p.%s.IsZero() {`, pqtfmt.Public(c.Name))
		braces++
	}

	g.Printf(strings.Replace(`
		if {{SELECTOR}}.Dirty {
			if _, err := {{SELECTOR}}.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := {{SELECTOR}}.WriteString(%s); err != nil {
			return "", nil, err
		}
		if _, err := {{SELECTOR}}.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := {{SELECTOR}}.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		{{SELECTOR}}.Add(p.%s)
		{{SELECTOR}}.Dirty=true
		`, "{{SELECTOR}}", sel, -1),
		pqtfmt.Public("table", c.Table.Name, "column", c.Name),
		pqtfmt.Public(c.Name),
	)

	if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
		if g.canBeNil(c, pqtgo.ModeOptional) || g.isNullable(c, pqtgo.ModeOptional) || g.isType(c, pqtgo.ModeOptional, "time.Time") {
			g.Printf(strings.Replace(`
				} else {
					if {{SELECTOR}}.Dirty {
						if _, err := {{SELECTOR}}.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if _, err := {{SELECTOR}}.WriteString(%s); err != nil {
						return "", nil, err
					}
					if _, err := {{SELECTOR}}.WriteString("=%s"); err != nil {
						return "", nil, err
					}
				{{SELECTOR}}.Dirty=true`, "{{SELECTOR}}", sel, -1),
				pqtfmt.Public("table", c.Table.Name, "column", c.Name),
				d,
			)
		}
	}

	closeBrace(g, braces)
}

func (g *Generator) scanJoinableRelationships(t *pqt.Table, sel string) {
	for _, r := range joinableRelationships(t) {
		joinPropertyName := pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name))
		nestedEntityPropertyName := pqtfmt.Public(or(r.InversedName, r.InversedTable.Name))

		if r.Type == pqt.RelationshipTypeOneToMany || r.Type == pqt.RelationshipTypeManyToMany {
			continue
		}
		g.Printf(`
			if %s.%s != nil && %s.%s.Kind.Actionable() && %s.%s.%s {
				ent.%s = &%sEntity{}
				if prop, err = ent.%s.%s(); err != nil {
					return nil, err
				}
				props = append(props, prop...)
			}`,
			sel,
			joinPropertyName,
			sel,
			joinPropertyName,
			sel,
			joinPropertyName,
			pqtfmt.Public("fetch"),
			nestedEntityPropertyName,
			pqtfmt.Public(r.InversedTable.Name),
			nestedEntityPropertyName,
			pqtfmt.Public("props"),
		)
	}
}

// entityPropertiesGenerator produces struct field definition for each column and relationship defined on a table.
// It thread differently relationship differently based on ownership.
func (g *Generator) entityPropertiesGenerator(t *pqt.Table) chan structField {
	fields := make(chan structField)

	go func(out chan structField) {
		for _, c := range t.Columns {
			if t := g.columnType(c, pqtgo.ModeDefault); t != "<nil>" {
				out <- structField{Name: pqtfmt.Public(c.Name), Type: t, ReadOnly: c.IsDynamic}
			}
		}

		for _, r := range t.OwnedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: pqtfmt.Public(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", pqtfmt.Public(r.InversedTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: pqtfmt.Public(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", pqtfmt.Public(r.InversedTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: pqtfmt.Public(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", pqtfmt.Public(r.InversedTable.Name))}
			}
		}

		for _, r := range t.InversedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: pqtfmt.Public(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", pqtfmt.Public(r.OwnerTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: pqtfmt.Public(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", pqtfmt.Public(r.OwnerTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: pqtfmt.Public(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", pqtfmt.Public(r.OwnerTable.Name))}
			}
		}

		for _, r := range t.ManyToManyRelationships {
			if r.Type != pqt.RelationshipTypeManyToMany {
				continue
			}

			switch {
			case r.OwnerTable == t:
				out <- structField{Name: pqtfmt.Public(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", pqtfmt.Public(r.InversedTable.Name))}
			case r.InversedTable == t:
				out <- structField{Name: pqtfmt.Public(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", pqtfmt.Public(r.OwnerTable.Name))}
			}
		}

		close(out)
	}(fields)

	return fields
}
