package pqtgogen

import (
	"fmt"
	"go/format"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

type Generator struct {
	Formatter  *Formatter
	Version    float64
	Pkg        string
	Imports    []string
	Plugins    []Plugin
	Components Component

	g *gogen.Generator
	p *print.Printer
}

// Generate ...
func (g *Generator) Generate(s *pqt.Schema) ([]byte, error) {
	if err := g.generate(s); err != nil {
		return nil, err
	}

	return format.Source(g.p.Bytes())
}

// GenerateTo ...
func (g *Generator) GenerateTo(w io.Writer, s *pqt.Schema) error {
	if err := g.generate(s); err != nil {
		return err
	}

	buf, err := format.Source(g.p.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func (g *Generator) generate(s *pqt.Schema) error {
	g.g = &gogen.Generator{}
	g.p = &g.g.Printer

	g.generatePackage()
	g.generateImports(s)
	if g.Components&ComponentRepository != 0 {
		g.generateLogFunc(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 || g.Components&ComponentHelpers != 0 {
		g.generateInterfaces(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
		g.generateRepositoryJoinClause(s)
	}
	for _, t := range s.Tables {
		g.generateConstantsAndVariables(t)
		g.generateEntity(t)
		g.generateEntityProp(t)
		g.generateEntityProps(t)
		if g.Components&ComponentHelpers != 0 {
			g.generateRepositoryScanRows(t)
		}
		if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
			g.generateIterator(t)
			g.generateCriteria(t)
			g.generateFindExpr(t)
			g.generateJoin(t)
		}
		if g.Components&ComponentCount != 0 {
			g.generateCountExpr(t)
		}
		if g.Components&ComponentUpdate != 0 || g.Components&ComponentUpsert != 0 {
			g.generatePatch(t)
		}
		if g.Components&ComponentRepository != 0 {
			g.generateRepository(t)
			if g.Components&ComponentInsert != 0 {
				g.generateRepositoryInsertQuery(t)
				g.generateRepositoryInsert(t)
			}
			if g.Components&ComponentFind != 0 {
				g.generateRepositoryWhereClause(t)
				g.generateRepositoryFindQuery(t)
				g.generateRepositoryFind(t)
				g.generateRepositoryFindIter(t)
				g.generateRepositoryFindOneByPrimaryKey(t)
				g.generateRepositoryFindOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpdate != 0 {
				g.generateRepositoryUpdateOneByPrimaryKeyQuery(t)
				g.generateRepositoryUpdateOneByPrimaryKey(t)
				g.generateRepositoryUpdateOneByUniqueConstraintQuery(t)
				g.generateRepositoryUpdateOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpsert != 0 {
				g.generateRepositoryUpsertQuery(t)
				g.generateRepositoryUpsert(t)
			}
			if g.Components&ComponentCount != 0 {
				g.generateRepositoryCount(t)
			}
			if g.Components&ComponentDelete != 0 {
				g.generateRepositoryDeleteOneByPrimaryKey(t)
			}
		}
	}
	g.generateStatics(s)

	return g.p.Err
}

func (g *Generator) generatePackage() {
	g.g.Package(g.Pkg)
}

func (g *Generator) generateImports(s *pqt.Schema) {
	g.g.Imports(s, "github.com/m4rw3r/uuid")
}

func (g *Generator) generateEntity(t *pqt.Table) {
	g.g.Entity(t)
	g.g.NewLine()
}

func (g *Generator) generateFindExpr(t *pqt.Table) {
	g.g.FindExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCountExpr(t *pqt.Table) {
	g.g.CountExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCriteria(t *pqt.Table) {
	g.g.Criteria(t)
	g.g.NewLine()
	g.g.Operand(t)
	g.g.NewLine()
}

func (g *Generator) generateJoin(t *pqt.Table) {
	g.g.Join(t)
	g.g.NewLine()
}

func (g *Generator) generatePatch(t *pqt.Table) {
	g.g.Patch(t)
	g.g.NewLine()
}

func (g *Generator) generateIterator(t *pqt.Table) {
	g.g.Iterator(t)
	g.g.NewLine()
}

func (g *Generator) generateRepository(t *pqt.Table) {
	g.g.Repository(t)
	g.g.NewLine()
}

func (g *Generator) generateConstantsAndVariables(t *pqt.Table) {
	g.g.Constraints(t)
	g.g.NewLine()
	g.g.Columns(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsertQuery(t *pqt.Table) {
	g.g.RepositoryInsertQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsert(t *pqt.Table) {
	g.g.RepositoryInsert(t)
	g.g.NewLine()
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
			if err = tmpl.Execute(g.p, map[string]interface{}{
				"selector": fmt.Sprintf("p.%s", g.Formatter.Identifier(c.Name)),
				"column":   g.Formatter.Identifier("table", c.Table.Name, "column", c.Name),
				"composer": sel,
			}); err != nil {
				panic(err)
			}
			g.p.Println("")
			return
		}
	}
	braces := 0
	if g.canBeNil(c, pqtgo.ModeOptional) {
		g.p.Printf(`
			if p.%s != nil {`, g.Formatter.Identifier(c.Name))
		braces++
	}
	if g.isNullable(c, pqtgo.ModeOptional) {
		g.p.Printf(`
			if p.%s.Valid {`, g.Formatter.Identifier(c.Name))
		braces++
	}
	if g.isType(c, pqtgo.ModeOptional, "time.Time") {
		g.p.Printf(`
			if !p.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
		braces++
	}

	g.p.Printf(strings.Replace(`
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
		g.Formatter.Identifier("table", c.Table.Name, "column", c.Name),
		g.Formatter.Identifier(c.Name),
	)

	if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
		if g.canBeNil(c, pqtgo.ModeOptional) || g.isNullable(c, pqtgo.ModeOptional) || g.isType(c, pqtgo.ModeOptional, "time.Time") {
			g.p.Printf(strings.Replace(`
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
				g.Formatter.Identifier("table", c.Table.Name, "column", c.Name),
				d,
			)
		}
	}

	closeBrace(g.p, braces)
}

func (g *Generator) generateRepositoryInsertClause(c *pqt.Column, sel string) {
	braces := 0

	switch c.Type {
	case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
		return
	default:
		if g.canBeNil(c, pqtgo.ModeDefault) {
			g.p.Printf(`
					if e.%s != nil {`,
				g.Formatter.Identifier(c.Name),
			)
			braces++
		}
		if g.isNullable(c, pqtgo.ModeDefault) {
			g.p.Printf(`
					if e.%s.Valid {`, g.Formatter.Identifier(c.Name))
			braces++
		}
		if g.isType(c, pqtgo.ModeDefault, "time.Time") {
			g.p.Printf(`
					if !e.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
			braces++
		}
		g.p.Printf(strings.Replace(`
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
			g.Formatter.Identifier("table", c.Table.Name, "column", c.Name),
			g.Formatter.Identifier(c.Name),
		)

		closeBrace(g.p, braces)
		g.p.Println("")
	}
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKeyQuery(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.p.Printf(`
		func (r *%sRepositoryBase) %sQuery(pk %s, p *%sPatch) (string, []interface{}, error) {`,
		entityName,
		g.Formatter.Identifier("UpdateOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.p.Printf(`
		buf := bytes.NewBufferString("UPDATE ")
		buf.WriteString(r.%s)
		update := NewComposer(%d)`,
		g.Formatter.Identifier("table"),
		len(t.Columns),
	)

	for _, c := range t.Columns {
		g.generateRepositorySetClause(c, "update")
	}
	g.p.Printf(`
	if !update.Dirty {
		return "", nil, errors.New("%s update failure, nothing to update")
	}`, entityName)

	g.p.Printf(`
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
		g.Formatter.Identifier("table", t.Name, "column", pk.Name),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("columns"),
	)

	g.p.Print(`
		buf.WriteString("`)
	g.selectList(t, -1)
	g.p.Print(`")
	}`)
	g.p.Print(`
		return buf.String(), update.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKey(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, g.Formatter.Identifier("updateOneBy", pk.Name), g.columnType(pk, pqtgo.ModeMandatory), entityName, entityName)
	g.p.Printf(`
		query, args, err := r.%sQuery(pk, p)
		if err != nil {
			return nil, err
		}`, g.Formatter.Identifier("updateOneBy", pk.Name))

	g.p.Printf(`
		var ent %sEntity
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(props...)`,
		entityName,
		g.Formatter.Identifier("props"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
	)
	g.p.Printf(`
		if r.%s != nil {
			r.%s(err, "%s", "update by primary key", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return &ent, nil
	}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraintQuery(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	for _, u := range g.uniqueConstraints(t) {
		method := []string{"updateOneBy"}
		arguments := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		method = append(method, "query")

		g.p.Printf(`
			func (r *%sRepositoryBase) %s(%s, p *%sPatch) (string, []interface{}, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
		)

		g.p.Printf(`
			buf := bytes.NewBufferString("UPDATE ")
			buf.WriteString(r.%s)
			update := NewComposer(%d)`, g.Formatter.Identifier("table"), len(u.PrimaryColumns))

		for _, c := range t.Columns {
			g.generateRepositorySetClause(c, "update")
		}
		g.p.Printf(`
			if !update.Dirty {
				return "", nil, errors.New("%s update failure, nothing to update")
			}`, entityName,
		)
		g.p.Print(`
			buf.WriteString(" SET ")
			buf.ReadFrom(update)
			buf.WriteString(" WHERE ")`)
		for i, c := range u.PrimaryColumns {
			if i != 0 {
				g.p.Print(`
					update.WriteString(" AND ")`)
			}
			g.p.Printf(`
				update.WriteString(%s)
				update.WriteString("=")
				update.WritePlaceholder()
				update.Add(%s)`,
				g.Formatter.Identifier("table", t.Name, "column", c.Name),
				g.Formatter.IdentifierPrivate(columnForeignName(c)),
			)
		}
		g.p.Printf(`
			buf.ReadFrom(update)
			buf.WriteString(" RETURNING ")
			if len(r.%s) > 0 {
				buf.WriteString(strings.Join(r.%s, ", "))
			} else {`,
			g.Formatter.Identifier("columns"),
			g.Formatter.Identifier("columns"),
		)

		g.p.Print(`
		buf.WriteString("`)
		g.selectList(t, -1)
		if len(u.Where) > 0 {
			g.p.Printf(` WHERE %s")
	}`, u.Where)
		} else {
			g.p.Print(`")
	}`)
		}
		g.p.Print(`
		return buf.String(), update.Args(), nil
	}`)
	}
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraint(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	for _, u := range g.uniqueConstraints(t) {
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
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
			arguments2 += g.Formatter.IdentifierPrivate(columnForeignName(c))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		g.p.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, %s, p *%sPatch) (*%sEntity, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
			entityName,
		)

		g.p.Printf(`
			query, args, err := r.%s(%s, p)
			if err != nil {
				return nil, err
			}`,
			g.Formatter.Identifier(append(method, "query")...),
			arguments2,
		)
		g.p.Printf(`
		var ent %sEntity
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(props...)`,
			entityName,
			g.Formatter.Identifier("props"),
			g.Formatter.Identifier("columns"),
			g.Formatter.Identifier("db"),
		)

		g.p.Printf(`
			if r.%s != nil {
				r.%s(err, "%s", "update one by unique", query, args...)
			}
			if err != nil {
				return nil, err
			}
			return &ent, nil
		}`,
			g.Formatter.Identifier("log"),
			g.Formatter.Identifier("log"),
			entityName,
		)
	}
}

func (g *Generator) generateRepositoryUpsertQuery(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}
	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, p *%sPatch, inf ...string) (string, []interface{}, error) {`,
		entityName,
		g.Formatter.Identifier("upsert"),
		entityName,
		entityName,
	)
	g.p.Printf(`
		upsert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns)*2, g.Formatter.Identifier("table"))

	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositoryInsertClause(c, "upsert")
	}

	g.p.Print(`
		if upsert.Dirty {
			buf.WriteString(" (")
			buf.ReadFrom(columns)
			buf.WriteString(") VALUES (")
			buf.ReadFrom(upsert)
			buf.WriteString(")")
		}
		buf.WriteString(" ON CONFLICT ")`,
	)

	g.p.Print(`
		if len(inf) > 0 {
		upsert.Dirty=false`)
	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositorySetClause(c, "upsert")
	}
	closeBrace(g.p, 1)

	g.p.Printf(`
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
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("columns"),
	)
	g.p.Print(`
		buf.WriteString("`)
	g.selectList(t, -1)
	g.p.Print(`")
	}`)
	g.p.Print(`
		}
		return buf.String(), upsert.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryUpsert(t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {`,
		entityName,
		g.Formatter.Identifier("upsert"),
		entityName,
		entityName,
		entityName,
	)
	g.p.Printf(`
			query, args, err := r.%sQuery(e, p, inf...)
			if err != nil {
				return nil, err
			}
			err = r.%s.QueryRowContext(ctx, query, args...).Scan(`,
		g.Formatter.Identifier("upsert"),
		g.Formatter.Identifier("db"),
	)

	for _, c := range t.Columns {
		g.p.Printf("&e.%s,\n", g.Formatter.Identifier(c.Name))
	}
	g.p.Printf(`)
		if r.%s != nil {
			r.%s(err, "%s", "upsert", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return e, nil
	}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
}

func (g *Generator) generateRepositoryWhereClause(t *pqt.Table) {
	name := g.Formatter.Identifier(t.Name)
	fnName := fmt.Sprintf("%sCriteriaWhereClause", name)
	g.p.Printf(`
		func %s(comp *Composer, c *%sCriteria, id int) (error) {`, fnName, name)

	g.p.Printf(`
	if c.child == nil {
		return _%s(comp, c, id)
	}
	node := c
	sibling := false
	for {
		if !sibling {
			if node.child != nil {
				if node.parent != nil {
					comp.WriteString("(")
				}
				node = node.child
				continue
			} else {
				comp.Dirty = false
				comp.WriteString("(")
				if err := _%s(comp, node, id); err != nil {
					return err
				}
				comp.WriteString(")")
			}
		}
		if node.sibling != nil {
			sibling = false
			comp.WriteString(" ")
			comp.WriteString(node.parent.operator)
			comp.WriteString(" ")
			node = node.sibling
			continue
		}
		if node.parent != nil {
			sibling = true
			if node.parent.parent != nil {
				comp.WriteString(")")
			}
			node = node.parent
			continue
		}

		break
	}
	return nil
	}`, fnName, fnName)

	g.p.Printf(`
		func _%sCriteriaWhereClause(comp *Composer, c *%sCriteria, id int) (error) {`, name, name)
ColumnsLoop:
	for _, c := range t.Columns {
		braces := 0
		for _, plugin := range g.Plugins {
			if txt := plugin.WhereClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(txt)
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(g.p, map[string]interface{}{
					"selector": fmt.Sprintf("c.%s", g.Formatter.Identifier(c.Name)),
					"column":   g.sqlSelector(c, "id"),
					"composer": "comp",
					"id":       "id",
				}); err != nil {
					panic(err)
				}
				g.p.Println("")
				continue ColumnsLoop
			}
		}
		if g.columnType(c, pqtgo.ModeCriteria) == "<nil>" {
			return
		}
		if g.canBeNil(c, pqtgo.ModeCriteria) {
			braces++
			g.p.Printf(`
				if c.%s != nil {`, g.Formatter.Identifier(c.Name))
		}
		if g.isNullable(c, pqtgo.ModeCriteria) {
			braces++
			g.p.Printf(`
				if c.%s.Valid {`, g.Formatter.Identifier(c.Name))
		}
		if g.isType(c, pqtgo.ModeCriteria, "time.Time") {
			braces++
			g.p.Printf(`
				if !c.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
		}

		g.p.Print(
			`if comp.Dirty {
				comp.WriteString(" AND ")
			}`)

		if c.IsDynamic {
			g.p.Printf(`
				if _, err := comp.WriteString("%s"); err != nil {
					return err
				}
				if _, err := comp.WriteString("("); err != nil {
					return err
				}`, c.Func.Name)
			for i, arg := range c.Func.Args {
				if arg.Type != c.Columns[i].Type {
					fmt.Printf("wrong function (%s) argument type, expected %v but got %v\n", c.Func.Name, arg.Type, c.Columns[i].Type)
				}
				if i != 0 {
					g.p.Print(`
					if _, err := comp.WriteString(", "); err != nil {
						return err
					}`)
				}
				g.p.Printf(`
					if err := comp.WriteAlias(id); err != nil {
						return err
					}
					if _, err := comp.WriteString(%s); err != nil {
						return err
					}`,
					g.Formatter.Identifier("table", c.Columns[i].Table.Name, "column", c.Columns[i].Name),
				)
			}
			g.p.Print(`
				if _, err := comp.WriteString(")"); err != nil {
					return err
				}`)
		} else {
			g.p.Printf(`
				if err := comp.WriteAlias(id); err != nil {
					return err
				}
				if _, err := comp.WriteString(%s); err != nil {
					return err
				}`,
				g.Formatter.Identifier("table", t.Name, "column", c.Name),
			)
		}

		g.p.Printf(`
			if _, err := comp.WriteString("="); err != nil {
				return err
			}
			if err := comp.WritePlaceholder(); err != nil {
				return err
			}
			comp.Add(c.%s)
			comp.Dirty=true`,
			g.Formatter.Identifier(c.Name),
		)
		closeBrace(g.p, braces)
		g.p.Println("")
	}
	g.p.Print(`
	return nil`)
	closeBrace(g.p, 1)
}

func (g *Generator) generateRepositoryJoinClause(s *pqt.Schema) {
	g.p.Print(`
	func joinClause(comp *Composer, jt JoinType, on string) (ok bool, err error) {
		if jt != JoinDoNot {
			switch jt {
			case JoinInner:
				if _, err = comp.WriteString(" INNER JOIN "); err != nil {
					return
				}
			case JoinLeft:
				if _, err = comp.WriteString(" LEFT JOIN "); err != nil {
					return
				}
			case JoinRight:
				if _, err = comp.WriteString(" RIGHT JOIN "); err != nil {
					return
				}
			case JoinCross:
				if _, err = comp.WriteString(" CROSS JOIN "); err != nil {
					return
				}
			default:
				return
			}
			if _, err = comp.WriteString(on); err != nil {
				return
			}
			comp.Dirty = true
			ok = true
			return
		}
		return
	}`,
	)
}

func (g *Generator) generateLogFunc(s *pqt.Schema) {
	g.p.Printf(`
	// %s represents function that can be passed into repository to log query result.
	type LogFunc func(err error, ent, fnc, sql string, args ...interface{})`,
		g.Formatter.Identifier("log", "func"),
	)
}

func (g *Generator) generateInterfaces(s *pqt.Schema) {
	g.p.Print(`
	// Rows ...
	type Rows interface {
		io.Closer
		ColumnTypes() ([]*sql.ColumnType, error)
		Columns() ([]string, error)
		Err() error
		Next() bool
		NextResultSet() bool
		Scan(dest ...interface{}) error
	}`)
}

func (g *Generator) selectList(t *pqt.Table, nb int) {
	for i, c := range t.Columns {
		if i != 0 {
			g.p.Print(", ")
		}
		if c.IsDynamic {
			g.p.Printf("%s(", c.Func.Name)
			for i, arg := range c.Func.Args {
				if arg.Type != c.Columns[i].Type {
					fmt.Printf("wrong function (%s) argument type, expected %v but got %v\n", c.Func.Name, arg.Type, c.Columns[i].Type)
				}
				if i != 0 {
					g.p.Print(", ")
				}
				if nb > -1 {
					g.p.Printf("t%d.%s", nb, c.Columns[i].Name)
				} else {
					g.p.Printf("%s", c.Columns[i].Name)
				}
			}
			g.p.Printf(") AS %s", c.Name)
		} else {
			if nb > -1 {
				g.p.Printf("t%d.%s", nb, c.Name)
			} else {
				g.p.Printf("%s", c.Name)
			}
		}
	}
}

func (g *Generator) scanJoinableRelationships(t *pqt.Table, sel string) {
	for _, r := range joinableRelationships(t) {
		if r.Type == pqt.RelationshipTypeOneToMany || r.Type == pqt.RelationshipTypeManyToMany {
			continue
		}
		g.p.Printf(`
			if %s.%s != nil && %s.%s.%s {
				ent.%s = &%sEntity{}
				if prop, err = ent.%s.%s(); err != nil {
					return nil, err
				}
				props = append(props, prop...)
			}`,
			sel,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			sel,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("fetch"),
			g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier(r.InversedTable.Name),
			g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("props"),
		)
	}
}

func (g *Generator) generateRepositoryFindQuery(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %sQuery(fe *%sFindExpr) (string, []interface{}, error) {`, entityName, g.Formatter.Identifier("find"), entityName)
	g.p.Printf(`
		comp := NewComposer(%d)
		buf := bytes.NewBufferString("SELECT ")
		if len(fe.%s) == 0 {
		buf.WriteString("`, len(t.Columns), g.Formatter.Identifier("columns"))
	g.selectList(t, 0)
	g.p.Printf(`")
		} else {
			buf.WriteString(strings.Join(fe.%s, ", "))
		}`, g.Formatter.Identifier("columns"))
	for nb, r := range joinableRelationships(t) {
		g.p.Printf(`
			if fe.%s != nil && fe.%s.%s {`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("fetch"),
		)
		g.p.Print(`
		buf.WriteString(", `)
		g.selectList(r.InversedTable, nb+1)
		g.p.Print(`")`)
		closeBrace(g.p, 1)
	}
	g.p.Printf(`
		buf.WriteString(" FROM ")
		buf.WriteString(r.%s)
		buf.WriteString(" AS t0")`, g.Formatter.Identifier("table"))
	for nb, r := range joinableRelationships(t) {
		oc := r.OwnerColumns
		ic := r.InversedColumns
		if len(oc) != len(ic) {
			panic("number of owned and inversed foreign key columns is not equal")
		}

		g.p.Printf(`
			if fe.%s != nil {`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
		)
		g.p.Printf(`
			joinClause(comp, fe.%s.%s, "%s AS t%d ON `,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("kind"),
			r.InversedTable.FullName(),
			nb+1,
		)

		for i := 0; i < len(oc); i++ {
			if i > 0 {
				g.p.Print(` AND `)
			}
			g.p.Printf(`t%d.%s=t%d.%s`, 0, oc[i].Name, nb+1, ic[i].Name)
		}
		g.p.Print(`")`)

		g.p.Printf(`
		if fe.%s.%s != nil {
			comp.Dirty = true
			if err := %sCriteriaWhereClause(comp, fe.%s.%s, %d); err != nil {
				return "", nil, err
			}
		}`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("on"),
			g.Formatter.Identifier(r.InversedTable.Name),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("on"),
			nb+1,
		)

		closeBrace(g.p, 1)
	}

	g.p.Printf(`
	if comp.Dirty {
		buf.ReadFrom(comp)
		comp.Dirty = false
	}
	if fe.%s != nil {
		if err := %sCriteriaWhereClause(comp, fe.%s, 0); err != nil {
			return "", nil, err
		}
	}`,
		g.Formatter.Identifier("where"),
		g.Formatter.Identifier(t.Name),
		g.Formatter.Identifier("where"),
	)

	for nb, r := range joinableRelationships(t) {
		g.p.Printf(`
		if fe.%s != nil && fe.%s.%s != nil {
			if err := %sCriteriaWhereClause(comp, fe.%s.%s, %d); err != nil {
				return "", nil, err
			}
		}`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("where"),
			g.Formatter.Identifier(r.InversedTable.Name),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("where"),
			nb+1,
		)
	}

	g.p.Print(`
		if comp.Dirty {
			if _, err := buf.WriteString(" WHERE "); err != nil {
				return "", nil, err
			}
			buf.ReadFrom(comp)
		}
	`)

	g.p.Printf(`
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
		g.Formatter.Identifier("orderBy"),
		g.Formatter.Identifier("orderBy"),
		g.Formatter.Identifier("table", t.Name, "columns"),
		g.Formatter.Identifier("offset"),
		g.Formatter.Identifier("offset"),
		g.Formatter.Identifier("limit"),
		g.Formatter.Identifier("limit"),
	)

	g.p.Print(`
		buf.ReadFrom(comp)
	`)
	g.p.Println(`
	return buf.String(), comp.Args(), nil
}`)
}

func (g *Generator) generateRepositoryFind(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) ([]*%sEntity, error) {`, entityName, g.Formatter.Identifier("find"), entityName, entityName)
	g.p.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("db"),
	)

	g.p.Printf(`
		if r.%s != nil {
			r.%s(err, "%s", "find", query, args...)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)

	g.p.Printf(`
		var entities []*%sEntity
		var props []interface{}
		for rows.Next() {
			var ent %sEntity
			if props, err = ent.%s(); err != nil {
				return nil, err
			}`,
		entityName,
		g.Formatter.Identifier(t.Name),
		g.Formatter.Identifier("props"),
	)
	if hasJoinableRelationships(t) {
		g.p.Print(`
		var prop []interface{}`)
	}
	g.scanJoinableRelationships(t, "fe")
	g.p.Print(`
			err = rows.Scan(props...)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &ent)
		}`)
	g.p.Printf(`
		err = rows.Err()
		if r.%s != nil {
			r.%s(err, "%s", "find", query, args...)
		}
		if err != nil {
			return nil, err
		}
		return entities, nil
	}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
}

func (g *Generator) generateRepositoryFindIter(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) (*%sIterator, error) {`, entityName, g.Formatter.Identifier("findIter"), entityName, entityName)
	g.p.Printf(`
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("db"),
	)

	g.p.Printf(`
	 	if r.%s != nil {
			r.%s(err, "%s", "find iter", query, args...)
		}
		if err != nil {
			return nil, err
		}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
	g.p.Printf(`
			return &%sIterator{
				rows: rows,
				expr: fe,
				cols: []string{`,
		g.Formatter.Identifier(t.Name),
	)
	for _, c := range t.Columns {
		g.p.Printf(`"%s",`, c.Name)
	}
	g.p.Print(`},
		}, nil
	}`)
}

func (g *Generator) generateRepositoryCount(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCountExpr) (int64, error) {`, entityName, g.Formatter.Identifier("count"), entityName)
	g.p.Printf(`
		query, args, err := r.%sQuery(&%sFindExpr{
			%s: c.%s,
			%s: []string{"COUNT(*)"},
		`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier(entityName),
		g.Formatter.Identifier("where"),
		g.Formatter.Identifier("where"),
		g.Formatter.Identifier("columns"),
	)
	for _, r := range joinableRelationships(t) {
		g.p.Printf(`
		%s: c.%s,`, g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)), g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)))
	}
	g.p.Printf(`
		})
		if err != nil {
			return 0, err
		}
		var count int64
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(&count)`,
		g.Formatter.Identifier("db"),
	)

	g.p.Printf(`
		if r.%s != nil {
			r.%s(err, "%s", "count", query, args...)
		}
		if err != nil {
			return 0, err
		}
		return count, nil
	}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (*%sEntity, error) {`,
		entityName,
		g.Formatter.Identifier("FindOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		entityName,
	)
	g.p.Printf(`
		find := NewComposer(%d)
		find.WriteString("SELECT ")
		if len(r.%s) == 0 {
			find.WriteString("`,
		len(t.Columns), g.Formatter.Identifier("columns"))
	g.selectList(t, -1)
	g.p.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, g.Formatter.Identifier("columns"))

	g.p.Printf(`
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
		g.Formatter.Identifier("table", t.Name),
		g.Formatter.Identifier("table", t.Name, "column", pk.Name),
		entityName,
	)

	g.p.Printf(`
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
		g.Formatter.Identifier("props"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
	)
	g.p.Printf(`
		if r.%s != nil {
			r.%s(err, "%s", "find by primary key", find.String(), find.Args()...)
		}
		if err != nil {
			return nil, err
		}
		return &ent, nil
	}`,
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("log"),
		entityName,
	)
}

func (g *Generator) generateRepositoryFindOneByUniqueConstraint(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	for _, u := range g.uniqueConstraints(t) {
		method := []string{"FindOneBy"}
		arguments := ""

		for i, c := range u.PrimaryColumns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, pqtgo.ModeMandatory))
		}

		if len(u.Where) > 0 && len(u.MethodSuffix) > 0 {
			method = append(method, "Where")
			method = append(method, u.MethodSuffix)
		}

		g.p.Printf(`
			func (r *%sRepositoryBase) %s(ctx context.Context, %s) (*%sEntity, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
		)
		g.p.Printf(`
			find := NewComposer(%d)
			find.WriteString("SELECT ")
					if len(r.%s) == 0 {
			find.WriteString("`,
			len(t.Columns), g.Formatter.Identifier("columns"))
		g.selectList(t, -1)
		g.p.Printf(`")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, g.Formatter.Identifier("columns"))

		partialClause := ""
		if len(u.Where) > 0 {
			partialClause = fmt.Sprintf("%s AND ", u.Where)
		}

		g.p.Printf(`
			find.WriteString(" FROM ")
			find.WriteString(%s)
			find.WriteString(" WHERE %s")`,
			g.Formatter.Identifier("table", t.Name),
			partialClause,
		)
		for i, c := range u.PrimaryColumns {
			if i != 0 {
				g.p.Print(`find.WriteString(" AND ")`)
			}
			g.p.Printf(`
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(%s)
		`, g.Formatter.Identifier("table", t.Name, "column", c.Name), g.Formatter.IdentifierPrivate(columnForeignName(c)))
		}

		g.p.Printf(`
			var (
				ent %sEntity
			)
			props, err := ent.%s(r.%s...)
			if err != nil {
				return nil, err
			}
			err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
			entityName,
			g.Formatter.Identifier("props"),
			g.Formatter.Identifier("columns"),
			g.Formatter.Identifier("db"),
		)
		g.p.Print(`
			if err != nil {
				return nil, err
			}

			return &ent, nil
		}`)
	}
}

func (g *Generator) generateRepositoryDeleteOneByPrimaryKey(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.p.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (int64, error) {`,
		entityName,
		g.Formatter.Identifier("DeleteOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
	)
	g.p.Printf(`
		find := NewComposer(%d)
		find.WriteString("DELETE FROM ")
		find.WriteString(%s)
		find.WriteString(" WHERE ")
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(pk)`, len(t.Columns),
		g.Formatter.Identifier("table", t.Name),
		g.Formatter.Identifier("table", t.Name, "column", pk.Name),
	)

	g.p.Printf(`
		res, err := r.%s.ExecContext(ctx, find.String(), find.Args()...)`,
		g.Formatter.Identifier("db"),
	)
	g.p.Print(`
		if err != nil {
				return 0, err
			}

		return res.RowsAffected()
	}`)
}

func (g *Generator) generateEntityProp(t *pqt.Table) {
	g.p.Printf(`
		func (e *%sEntity) %s(cn string) (interface{}, bool) {`, g.Formatter.Identifier(t.Name), g.Formatter.Identifier("prop"))
	g.p.Println(`
		switch cn {`)

ColumnsLoop:
	for _, c := range t.Columns {
		g.p.Printf(`
			case %s:`, g.Formatter.Identifier("table", t.Name, "column", c.Name))
		for _, plugin := range g.Plugins {
			if txt := plugin.ScanClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(fmt.Sprintf(`
					return %s, true`, txt))
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(g.p, map[string]interface{}{
					"selector": fmt.Sprintf("e.%s", g.Formatter.Identifier(c.Name)),
				}); err != nil {
					panic(err)
				}
				g.p.Println("")
				continue ColumnsLoop
			}
		}
		switch {
		case g.isArray(c, pqtgo.ModeDefault):
			pn := g.Formatter.Identifier(c.Name)
			switch g.columnType(c, pqtgo.ModeDefault) {
			case "pq.Int64Array":
				g.p.Printf(`if e.%s == nil { e.%s = []int64{} }`, pn, pn)
			case "pq.StringArray":
				g.p.Printf(`if e.%s == nil { e.%s = []string{} }`, pn, pn)
			case "pq.Float64Array":
				g.p.Printf(`if e.%s == nil { e.%s = []float64{} }`, pn, pn)
			case "pq.BoolArray":
				g.p.Printf(`if e.%s == nil { e.%s = []bool{} }`, pn, pn)
			case "pq.ByteaArray":
				g.p.Printf(`if e.%s == nil { e.%s = [][]byte{} }`, pn, pn)
			}

			g.p.Printf(`
				return &e.%s, true`, g.Formatter.Identifier(c.Name))
		case g.canBeNil(c, pqtgo.ModeDefault):
			g.p.Printf(`
				return e.%s, true`,
				g.Formatter.Identifier(c.Name),
			)
		default:
			g.p.Printf(`
				return &e.%s, true`,
				g.Formatter.Identifier(c.Name),
			)
		}
	}

	g.p.Print(`
		default:`)
	g.p.Print(`
		return nil, false`)
	g.p.Print("}\n}\n")

}

func (g *Generator) generateEntityProps(t *pqt.Table) {
	g.p.Printf(`
		func (e *%sEntity) %s(cns ...string) ([]interface{}, error) {`, g.Formatter.Identifier(t.Name), g.Formatter.Identifier("props"))
	g.p.Printf(`
		if len(cns) == 0 {
			cns = %s
		}
		res := make([]interface{}, 0, len(cns))
		for _, cn := range cns {
			if prop, ok := e.%s(cn); ok {
				res = append(res, prop)
			} else {
				return nil, fmt.Errorf("unexpected column provided: %%s", cn)
			}
		}
		return res, nil`,
		g.Formatter.Identifier("table", t.Name, "columns"),
		g.Formatter.Identifier("prop"),
	)
	g.p.Print(`
		}`)
}

func (g *Generator) generateRepositoryScanRows(t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	funcName := g.Formatter.Identifier("scan", t.Name, "rows")
	g.p.Printf(`
		// %s helps to scan rows straight to the slice of entities.
		func %s(rows Rows) (entities []*%sEntity, err error) {`, funcName, funcName, entityName)
	g.p.Printf(`
		for rows.Next() {
			var ent %sEntity
			err = rows.Scan(
			`, entityName,
	)
	for _, c := range t.Columns {
		g.p.Printf("&ent.%s,\n", g.Formatter.Identifier(c.Name))
	}
	g.p.Print(`)
			if err != nil {
				return
			}

			entities = append(entities, &ent)
		}
		if err = rows.Err(); err != nil {
			return
		}

		return
	}`)
}

func (g *Generator) isArray(c *pqt.Column, m int32) bool {
	if strings.HasPrefix(g.columnType(c, m), "[]") {
		return true
	}

	return g.isType(c, m, "pq.StringArray", "pq.Int64Array", "pq.BoolArray", "pq.Float64Array", "pq.ByteaArray", "pq.GenericArray")
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

func (g *Generator) columnType(c *pqt.Column, m int32) string {
	m = columnMode(c, m)
	for _, plugin := range g.Plugins {
		if txt := plugin.PropertyType(c, m); txt != "" {
			return txt
		}
	}
	return g.Formatter.Type(c.Type, m)
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

func (g *Generator) generateStatics(s *pqt.Schema) {
	g.p.Print(`

const (
	JoinDoNot = iota
	JoinInner
	JoinLeft
	JoinRight
	JoinCross
)

type JoinType int

func (jt JoinType) String() string {
	switch jt {

	case JoinInner:
		return "INNER JOIN"
	case JoinLeft:
		return "LEFT JOIN"
	case JoinRight:
		return "RIGHT JOIN"
	case JoinCross:
		return "CROSS JOIN"
	default:
		return ""
	}
}

// ErrorConstraint returns the error constraint of err if it was produced by the pq library.
// Otherwise, it returns empty string.
func ErrorConstraint(err error) string {
	if err == nil {
		return ""
	}
	if pqerr, ok := err.(*pq.Error); ok {
		return pqerr.Constraint
	}

	return ""
}

type RowOrder struct {
	Name string
	Descending bool
}

type NullInt64Array struct {
	pq.Int64Array
	Valid  bool
}

func (n *NullInt64Array) Scan(value interface{}) error {
	if value == nil {
		n.Int64Array, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.Int64Array.Scan(value)
}

type NullFloat64Array struct {
	pq.Float64Array
	Valid  bool
}

func (n *NullFloat64Array) Scan(value interface{}) error {
	if value == nil {
		n.Float64Array, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.Float64Array.Scan(value)
}

type NullBoolArray struct {
	pq.BoolArray
	Valid  bool
}

func (n *NullBoolArray) Scan(value interface{}) error {
	if value == nil {
		n.BoolArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.BoolArray.Scan(value)
}

type NullStringArray struct {
	pq.StringArray
	Valid  bool
}

func (n *NullStringArray) Scan(value interface{}) error {
	if value == nil {
		n.StringArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.StringArray.Scan(value)
}

type NullByteaArray struct {
	pq.ByteaArray
	Valid  bool
}

func (n *NullByteaArray) Scan(value interface{}) error {
	if value == nil {
		n.ByteaArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.ByteaArray.Scan(value)
}


const (
	jsonArraySeparator     = ","
	jsonArrayBeginningChar = "["
	jsonArrayEndChar       = "]"
)

// JSONArrayInt64 is a slice of int64s that implements necessary interfaces.
type JSONArrayInt64 []int64

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayInt64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayInt64, 0)
		}
		return nil
	}

	var tmp []string
	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
	}

	if l == 2 {
		*a = make(JSONArrayInt64, 0)
		return nil
	}

	if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
		return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
	*a = make(JSONArrayInt64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s at index %d", v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayInt64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatInt(v, 10)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// JSONArrayString is a slice of strings that implements necessary interfaces.
type JSONArrayString []string

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayString) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayString, 0)
		}
		return nil
	}

	switch t := src.(type) {
	case []byte:
		return json.Unmarshal(t, a)
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayString) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// JSONArrayFloat64 is a slice of int64s that implements necessary interfaces.
type JSONArrayFloat64 []float64

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayFloat64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayFloat64, 0)
		}
		return nil
	}

	var tmp []string
	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
	}

	if l == 2 {
		*a = make(JSONArrayFloat64, 0)
		return nil
	}

	if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
		return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
	*a = make(JSONArrayFloat64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s at index %d", v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayFloat64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}


var (
	// Space is a shorthand composition option that holds space.
	Space = &CompositionOpts{
		Joint: " ",
	}
	// And is a shorthand composition option that holds AND operator.
	And = &CompositionOpts{
		Joint: " AND ",
	}
	// Or is a shorthand composition option that holds OR operator.
	Or = &CompositionOpts{
		Joint: " OR ",
	}
	// Comma is a shorthand composition option that holds comma.
	Comma = &CompositionOpts{
		Joint: ", ",
	}
)

// CompositionOpts is a container for modification that can be applied.
type CompositionOpts struct {
	Joint                         string
	PlaceholderFuncs, SelectorFuncs []string
	PlaceholderCast, SelectorCast   string
	IsJSON                        bool
	IsDynamic                     bool
}

// CompositionWriter is a simple wrapper for WriteComposition function.
type CompositionWriter interface {
	// WriteComposition is a function that allow custom struct type to be used as a part of criteria.
	// It gives possibility to write custom query based on object that implements this interface.
	WriteComposition(string, *Composer, *CompositionOpts) error
}

// Composer holds buffer, arguments and placeholders count.
// In combination with external buffet can be also used to also generate sub-queries.
// To do that simply write buffer to the parent buffer, composer will hold all arguments and remember number of last placeholder.
type Composer struct {
	buf     bytes.Buffer
	args    []interface{}
	counter int
	Dirty   bool
}

// NewComposer allocates new Composer with inner slice of arguments of given size.
func NewComposer(size int64) *Composer {
	return &Composer{
		counter: 1,
		args:    make([]interface{}, 0, size),
	}
}

// WriteString appends the contents of s to the query buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with bytes ErrTooLarge.
func (c *Composer) WriteString(s string) (int, error) {
	return c.buf.WriteString(s)
}

// Write implements io Writer interface.
func (c *Composer) Write(b []byte) (int, error) {
	return c.buf.Write(b)
}

// Read implements io Reader interface.
func (c *Composer) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}

// ResetBuf resets internal buffer.
func (c *Composer) ResetBuf() {
	c.buf.Reset()
}

// String implements fmt Stringer interface.
func (c *Composer) String() string {
	return c.buf.String()
}

// WritePlaceholder writes appropriate placeholder to the query buffer based on current state of the composer.
func (c *Composer) WritePlaceholder() error {
	if _, err := c.buf.WriteString("$"); err != nil {
		return err
	}
	if _, err := c.buf.WriteString(strconv.Itoa(c.counter)); err != nil {
		return err
	}

	c.counter++
	return nil
}

func (c *Composer) WriteAlias(i int) error {
	if i < 0 {
		return nil
	}
	if _, err := c.buf.WriteString("t"); err != nil {
		return err
	}
	if _, err := c.buf.WriteString(strconv.Itoa(i)); err != nil {
		return err
	}
	if _, err := c.buf.WriteString("."); err != nil {
		return err
	}
	return nil
}

// Len returns number of arguments.
func (c *Composer) Len() int {
	return c.counter
}

// Add appends list with new element.
func (c *Composer) Add(arg interface{}) {
	c.args = append(c.args, arg)
}

// Args returns all arguments stored as a slice.
func (c *Composer) Args() []interface{} {
	return c.args
}`)

	for _, plugin := range g.Plugins {
		if txt := plugin.Static(s); txt != "" {
			g.p.Print(txt)
			g.p.Print("\n\n")
		}
	}
}

func (g *Generator) uniqueConstraints(t *pqt.Table) []*pqt.Constraint {
	var unique []*pqt.Constraint
	for _, c := range t.Constraints {
		if c.Type == pqt.ConstraintTypeUnique || c.Type == pqt.ConstraintTypeUniqueIndex {
			unique = append(unique, c)
		}
	}
	if len(unique) < 1 {
		return nil
	}
	return unique
}

func joinableRelationships(t *pqt.Table) (rels []*pqt.Relationship) {
	for _, r := range t.OwnedRelationships {
		if r.Type == pqt.RelationshipTypeOneToMany || r.Type == pqt.RelationshipTypeManyToMany {
			continue
		}
		rels = append(rels, r)
	}
	return
}

func hasJoinableRelationships(t *pqt.Table) bool {
	return len(joinableRelationships(t)) > 0
}

func (g *Generator) sqlSelector(c *pqt.Column, id string) string {
	if !c.IsDynamic {
		return g.Formatter.Identifier("table", c.Table.Name, "column", c.Name)
	}
	sel := c.Func.Name
	sel += "("
	for i := range c.Func.Args {
		if i != 0 {
			sel += ", "
		}
		sel += fmt.Sprint("t%d.")
		sel += c.Columns[i].Name
	}
	sel += ")"

	ret := fmt.Sprintf(`fmt.Sprintf("%s"`, sel)
	for range c.Func.Args {
		ret += ", "
		ret += id
	}
	return ret + ")"
}
