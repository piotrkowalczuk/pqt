package pqtgo

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/piotrkowalczuk/pqt"
)

const (
	ModeDefault = iota
	ModeMandatory
	ModeOptional
	ModeCriteria

	// Public ...
	Public Visibility = "public"
	// Private ...
	Private Visibility = "private"
)

// Visibility ...
type Visibility string

type Formatter struct {
	Visibility Visibility
	Acronyms   map[string]string
}

func (f *Formatter) Identifier(args ...string) (r string) {
	switch len(args) {
	case 0:
	case 1:
		r = f.identifier(args[0], f.Visibility)
	default:
		r = f.identifier(args[0], f.Visibility)
		for _, s := range args[1:] {
			r += f.identifier(s, Public)
		}
	}
	return r
}

func (f *Formatter) IdentifierPrivate(args ...string) (r string) {
	switch len(args) {
	case 0:
	case 1:
		r = f.identifier(args[0], Private)
	default:
		r = f.identifier(args[0], Private)
		for _, s := range args[1:] {
			r += f.identifier(s, Public)
		}
	}
	return r
}

func (f *Formatter) identifier(s string, v Visibility) string {
	r := snake(s, v == Private, f.Acronyms)
	if a, ok := keywords[r]; ok {
		return a
	}
	return r
}

func (f *Formatter) Type(t pqt.Type, m int32) string {
	switch tt := t.(type) {
	case pqt.MappableType:
		for _, mt := range tt.Mapping {
			return f.Type(mt, m)
		}
		return ""
	case BuiltinType:
		return generateTypeBuiltin(tt, m)
	case pqt.BaseType:
		return generateTypeBase(tt, m)
	case CustomType:
		return generateCustomType(tt, m)
	default:
		return ""
	}
}

type Generator struct {
	Formatter *Formatter
	Version   float64
	Pkg       string
	Imports   []string
	Plugins   []Plugin
}

// Generate ...
func (g *Generator) Generate(s *pqt.Schema) ([]byte, error) {
	code, err := g.generate(s)
	if err != nil {
		return nil, err
	}

	return code.Bytes(), nil
}

// GenerateTo ...
func (g *Generator) GenerateTo(w io.Writer, s *pqt.Schema) error {
	code, err := g.generate(s)
	if err != nil {
		return err
	}

	_, err = code.WriteTo(w)
	return err
}

func (g *Generator) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	b := bytes.NewBuffer(nil)

	g.generatePackage(b)
	g.generateImports(b, s)
	g.generateRepositoryJoinClause(b, s)
	for _, t := range s.Tables {
		g.generateConstants(b, t)
		g.generateColumns(b, t)
		g.generateEntity(b, t)
		g.generateEntityProp(b, t)
		g.generateEntityProps(b, t)
		g.generateIterator(b, t)
		g.generateCriteria(b, t)
		g.generateFindExpr(b, t)
		g.generateCountExpr(b, t)
		g.generateJoin(b, t)
		g.generatePatch(b, t)
		g.generateRepositoryScanRows(b, t)
		g.generateRepository(b, t)
		g.generateRepositoryInsertQuery(b, t)
		g.generateRepositoryInsert(b, t)
		g.generateRepositoryWhereClause(b, t)
		g.generateRepositoryFindQuery(b, t)
		g.generateRepositoryFind(b, t)
		g.generateRepositoryFindIter(b, t)
		g.generateRepositoryFindOneByPrimaryKey(b, t)
		g.generateRepositoryFindOneByUniqueConstraint(b, t)
		g.generateRepositoryUpdateOneByPrimaryKeyQuery(b, t)
		g.generateRepositoryUpdateOneByPrimaryKey(b, t)
		g.generateRepositoryUpdateOneByUniqueConstraintQuery(b, t)
		g.generateRepositoryUpdateOneByUniqueConstraint(b, t)
		g.generateRepositoryUpsertQuery(b, t)
		g.generateRepositoryUpsert(b, t)
		g.generateRepositoryCount(b, t)
		g.generateRepositoryDeleteOneByPrimaryKey(b, t)
	}
	g.generateStatics(b, s)

	return b, nil
}

func (g *Generator) generatePackage(w io.Writer) {
	pkg := g.Pkg
	if pkg == "" {
		pkg = "main"
	}
	fmt.Fprintf(w, "package %s\n", pkg)
}

func (g *Generator) generateImports(w io.Writer, schema *pqt.Schema) {
	imports := []string{
		"github.com/go-kit/kit/log",
		"github.com/m4rw3r/uuid",
	}
	imports = append(imports, g.Imports...)
	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(CustomType); ok {
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
			}
		}
	}

	fmt.Fprintln(w, "import(")
	for _, imp := range imports {
		io.WriteString(w, `"`)
		fmt.Fprint(w, imp)
		fmt.Fprintln(w, `"`)
	}
	fmt.Fprintln(w, ")")
}

func (g *Generator) generateEntity(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		// %sEntity ...`, g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		type %sEntity struct{`, g.Formatter.Identifier(t.Name))
	for prop := range g.entityPropertiesGenerator(t) {
		fmt.Fprintf(w, `
			// %s ...`, g.Formatter.Identifier(prop.Name))
		if prop.ReadOnly {
			fmt.Fprintf(w, `
			// %s is read only`, g.Formatter.Identifier(prop.Name))
		}
		if prop.Tags != "" {
			fmt.Fprintf(w,
				`%s %s %s`,
				g.Formatter.Identifier(prop.Name), prop.Type, prop.Tags,
			)
		} else {
			fmt.Fprintf(w, `
				%s %s`,
				g.Formatter.Identifier(prop.Name),
				prop.Type,
			)
		}
	}
	fmt.Fprint(w, `}`)
}

func (g *Generator) generateFindExpr(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		type %sFindExpr struct {`, g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		%s *%sCriteria`, g.Formatter.Identifier("where"), g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		%s, %s int64`, g.Formatter.Identifier("offset"), g.Formatter.Identifier("limit"))
	fmt.Fprintf(w, `
		%s []string`, g.Formatter.Identifier("columns"))
	fmt.Fprintf(w, `
		%s map[string]bool`, g.Formatter.Identifier("orderBy"))
	for _, r := range joinableRelationships(t) {
		fmt.Fprintf(w, `
		%s *%sJoin`, g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)), g.Formatter.Identifier(r.InversedTable.Name))
	}
	fmt.Fprintln(w, `
		}`)
}

func (g *Generator) generateCountExpr(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		type %sCountExpr struct {`, g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		%s *%sCriteria`, g.Formatter.Identifier("where"), g.Formatter.Identifier(t.Name))
	for _, r := range joinableRelationships(t) {
		fmt.Fprintf(w, `
		%s *%sJoin`, g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)), g.Formatter.Identifier(r.InversedTable.Name))
	}
	fmt.Fprintln(w, `
		}`)
}

func (g *Generator) generateCriteria(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		type %sCriteria struct {`, g.Formatter.Identifier(t.Name))
	for _, c := range t.Columns {
		if t := g.columnType(c, ModeCriteria); t != "<nil>" {
			fmt.Fprintf(w, `
				%s %s`, g.Formatter.Identifier(c.Name), t)
		}
	}
	fmt.Fprintln(w, `
		}`)
}

func (g *Generator) generateJoin(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		type %sJoin struct {`, g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		%s, %s *%sCriteria`, g.Formatter.Identifier("on"), g.Formatter.Identifier("where"), g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, `
		%s bool`, g.Formatter.Identifier("fetch"))
	fmt.Fprintf(w, `
		%s JoinType`, g.Formatter.Identifier("kind"))
	for _, r := range joinableRelationships(t) {
		fmt.Fprintf(w, `
		Join%s *%sJoin`, g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)), g.Formatter.Identifier(r.InversedTable.Name))
	}
	fmt.Fprintln(w, `
		}`)
}

func (g *Generator) generatePatch(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		type %sPatch struct {`, g.Formatter.Identifier(t.Name))

ArgumentsLoop:
	for _, c := range t.Columns {
		if c.PrimaryKey {
			continue ArgumentsLoop
		}

		if t := g.columnType(c, ModeOptional); t != "<nil>" {
			fmt.Fprintf(w, `
				%s %s`,
				g.Formatter.Identifier(c.Name),
				t,
			)
		}
	}
	fmt.Fprintln(w, `
		}`)
}

func (g *Generator) generateIterator(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	fmt.Fprintf(w, `

// %sIterator is not thread safe.
type %sIterator struct {
	rows *sql.Rows
	cols []string
}

func (i *%sIterator) Next() bool {
	return i.rows.Next()
}

func (i *%sIterator) Close() error {
	return i.rows.Close()
}

func (i *%sIterator) Err() error {
	return i.rows.Err()
}

// Columns is wrapper around sql.Rows.Columns method, that also cache output inside iterator.
func (i *%sIterator) Columns() ([]string, error) {
	if i.cols == nil {
		cols, err := i.rows.Columns()
		if err != nil {
			return nil, err
		}
		i.cols = cols
	}
	return i.cols, nil
}

// Ent is wrapper around %s method that makes iterator more generic.
func (i *%sIterator) Ent() (interface{}, error) {
	return i.%s()
}

func (i *%sIterator) %s() (*%sEntity, error) {
	var ent %sEntity
	cols, err := i.Columns()
	if err != nil {
		return nil, err
	}

	props, err := ent.%s(cols...)
	if err != nil {
		return nil, err
	}
	if err := i.rows.Scan(props...); err != nil {
		return nil, err
	}
	return &ent, nil
}
`, entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		g.Formatter.Identifier(t.Name),
		entityName,
		g.Formatter.Identifier(t.Name),
		entityName,
		entityName,
		g.Formatter.Identifier("props"),
	)
}

// entityPropertiesGenerator produces struct field definition for each column and relationship defined on a table.
// It thread differently relationship differently based on ownership.
func (g *Generator) entityPropertiesGenerator(t *pqt.Table) chan structField {
	fields := make(chan structField)

	go func(out chan structField) {
		for _, c := range t.Columns {
			if t := g.columnType(c, ModeDefault); t != "<nil>" {
				out <- structField{Name: g.Formatter.Identifier(c.Name), Type: t, ReadOnly: c.IsDynamic}
			}
		}

		for _, r := range t.OwnedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.Formatter.Identifier(r.InversedTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", g.Formatter.Identifier(r.InversedTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", g.Formatter.Identifier(r.InversedTable.Name))}
			}
		}

		for _, r := range t.InversedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: g.Formatter.Identifier(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", g.Formatter.Identifier(r.OwnerTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: g.Formatter.Identifier(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", g.Formatter.Identifier(r.OwnerTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: g.Formatter.Identifier(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.Formatter.Identifier(r.OwnerTable.Name))}
			}
		}

		for _, r := range t.ManyToManyRelationships {
			if r.Type != pqt.RelationshipTypeManyToMany {
				continue
			}

			switch {
			case r.OwnerTable == t:
				out <- structField{Name: g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.Formatter.Identifier(r.InversedTable.Name))}
			case r.InversedTable == t:
				out <- structField{Name: g.Formatter.Identifier(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.Formatter.Identifier(r.OwnerTable.Name))}
			}
		}

		close(out)
	}(fields)

	return fields
}

func (g *Generator) generateRepository(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `
		type %sRepositoryBase struct {
			%s string
			%s []string
			%s *sql.DB
			%s bool
			%s log.Logger
		}`,
		g.Formatter.Identifier(table.Name),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
	)
}

func (g *Generator) generateColumns(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `
		var (
			%s  = []string{`, g.Formatter.Identifier("table", table.Name, "columns"))

	for _, c := range table.Columns {
		fmt.Fprintf(w, `
			%s,`, g.Formatter.Identifier("table", table.Name, "column", c.Name))
	}
	io.WriteString(w, `
		})`)
}

func (g *Generator) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString(`
		const (`)
	g.generateConstantsColumns(code, table)
	g.generateConstantsConstraints(code, table)
	code.WriteString(`
		)`)
}

func (g *Generator) generateConstantsColumns(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `
		%s = "%s"`, g.Formatter.Identifier("table", table.Name), table.FullName())

	for _, c := range table.Columns {
		fmt.Fprintf(w, `
			%s = "%s"`, g.Formatter.Identifier("table", table.Name, "column", c.Name), c.Name)
	}
}

func (g *Generator) generateConstantsConstraints(w io.Writer, table *pqt.Table) {
	for _, c := range tableConstraints(table) {
		name := fmt.Sprintf(`%s`, pqt.JoinColumns(c.Columns, "_"))
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "Check"), c.String())
		case pqt.ConstraintTypePrimaryKey:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraintPrimaryKey"), c.String())
		case pqt.ConstraintTypeForeignKey:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "ForeignKey"), c.String())
		case pqt.ConstraintTypeExclusion:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "Exclusion"), c.String())
		case pqt.ConstraintTypeUnique:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "Unique"), c.String())
		case pqt.ConstraintTypeIndex:
			fmt.Fprintf(w, `
				%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "Index"), c.String())
		}

		io.WriteString(w, `
			`)
	}
}

func (g *Generator) generateRepositoryInsertQuery(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %sQuery(e *%sEntity) (string, []interface{}, error) {`, entityName, g.Formatter.Identifier("insert"), entityName)
	fmt.Fprintf(w, `
		insert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns), g.Formatter.Identifier("table"))

	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositoryInsertClause(w, c, "insert")
	}
	fmt.Fprint(w, `
		if columns.Len() > 0 {
			buf.WriteString(" (")
			buf.ReadFrom(columns)
			buf.WriteString(") VALUES (")
			buf.ReadFrom(insert)
			buf.WriteString(") ")`)
	fmt.Fprintf(w, `
			buf.WriteString("RETURNING ")
			if len(r.%s) > 0 {
				buf.WriteString(strings.Join(r.%s, ", "))
			} else {`,
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("columns"),
	)
	fmt.Fprint(w, `
		buf.WriteString("`)
	selectList(w, t, -1)
	fmt.Fprint(w, `")
	}`)
	fmt.Fprint(w, `
		}
		return buf.String(), insert.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryInsert(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity) (*%sEntity, error) {`, entityName, g.Formatter.Identifier("insert"), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(e)
			if err != nil {
				return nil, err
			}
			if err := r.%s.QueryRowContext(ctx, query, args...).Scan(`,
		g.Formatter.Identifier("insert"),
		g.Formatter.Identifier("db"),
	)

	for _, c := range table.Columns {
		fmt.Fprintf(w, "&e.%s,\n", g.Formatter.Identifier(c.Name))
	}
	fmt.Fprintf(w, `); err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return nil, err
		}
		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query success", "query", query, "table", r.%s)
		}
		return e, nil
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositorySetClause(w io.Writer, c *pqt.Column, sel string) {
	if c.PrimaryKey {
		return
	}
	for _, plugin := range g.Plugins {
		if txt := plugin.SetClause(c); txt != "" {
			tmpl, err := template.New("root").Parse(txt)
			if err != nil {
				panic(err)
			}
			if err = tmpl.Execute(w, map[string]interface{}{
				"selector": fmt.Sprintf("p.%s", g.Formatter.Identifier(c.Name)),
				"column":   g.Formatter.Identifier("table", c.Table.Name, "column", c.Name),
				"composer": sel,
			}); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, "")
			return
		}
	}
	braces := 0
	if g.canBeNil(c, ModeOptional) {
		fmt.Fprintf(w, `
			if p.%s != nil {`, g.Formatter.Identifier(c.Name))
		braces++
	}
	if g.isNullable(c, ModeOptional) {
		fmt.Fprintf(w, `
			if p.%s.Valid {`, g.Formatter.Identifier(c.Name))
		braces++
	}
	if g.isType(c, ModeOptional, "time.Time") {
		fmt.Fprintf(w, `
			if !p.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
		braces++
	}

	fmt.Fprintf(w, strings.Replace(`
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
		switch c.Type {
		case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
			fmt.Fprintf(w, strings.Replace(`
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

	closeBrace(w, braces)
}

func (g *Generator) generateRepositoryInsertClause(w io.Writer, c *pqt.Column, sel string) {
	braces := 0

	switch c.Type {
	case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
		return
	default:
		if g.canBeNil(c, ModeDefault) {
			fmt.Fprintf(w, `
					if e.%s != nil {`,
				g.Formatter.Identifier(c.Name),
			)
			braces++
		}
		if g.isNullable(c, ModeDefault) {
			fmt.Fprintf(w, `
					if e.%s.Valid {`, g.Formatter.Identifier(c.Name))
			braces++
		}
		if g.isType(c, ModeDefault, "time.Time") {
			fmt.Fprintf(w, `
					if !e.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
			braces++
		}
		fmt.Fprintf(w, strings.Replace(`
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

		closeBrace(w, braces)
		fmt.Fprintln(w, "")
	}
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKeyQuery(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %sQuery(pk %s, p *%sPatch) (string, []interface{}, error) {`,
		entityName,
		g.Formatter.Identifier("UpdateOneBy", pk.Name),
		g.columnType(pk, ModeMandatory),
		entityName,
	)
	fmt.Fprintf(w, `
		buf := bytes.NewBufferString("UPDATE ")
		buf.WriteString(r.%s)
		update := NewComposer(%d)`,
		g.Formatter.Identifier("table"),
		len(t.Columns),
	)

	for _, c := range t.Columns {
		g.generateRepositorySetClause(w, c, "update")
	}
	fmt.Fprintf(w, `
	if !update.Dirty {
		return "", nil, errors.New("%s update failure, nothing to update")
	}`, entityName)

	fmt.Fprintf(w, `
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

	fmt.Fprint(w, `
		buf.WriteString("`)
	selectList(w, t, -1)
	fmt.Fprint(w, `")
	}`)
	fmt.Fprint(w, `
		return buf.String(), update.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, g.Formatter.Identifier("updateOneBy", pk.Name), g.columnType(pk, ModeMandatory), entityName, entityName)
	fmt.Fprintf(w, `
		query, args, err := r.%sQuery(pk, p)
		if err != nil {
			return nil, err
		}`, g.Formatter.Identifier("updateOneBy", pk.Name))

	fmt.Fprintf(w, `
		var ent %sEntity
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		if err = r.%s.QueryRowContext(ctx, query, args...).Scan(props...)`,
		entityName,
		g.Formatter.Identifier("props"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
	)
	fmt.Fprintf(w, `; err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "update by primary key query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return nil, err
		}
		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "update by primary key query success", "query", query, "table", r.%s)
		}
		return &ent, nil
	}`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraintQuery(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	for _, u := range g.uniqueConstraints(table) {
		method := []string{"updateOneBy"}
		arguments := ""

		for i, c := range u.Columns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, ModeMandatory))
		}
		method = append(method, "query")

		fmt.Fprintf(w, `
			func (r *%sRepositoryBase) %s(%s, p *%sPatch) (string, []interface{}, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
		)

		fmt.Fprintf(w, `
			buf := bytes.NewBufferString("UPDATE ")
			buf.WriteString(r.%s)
			update := NewComposer(%d)`, g.Formatter.Identifier("table"), len(u.Columns))

		for _, c := range table.Columns {
			g.generateRepositorySetClause(w, c, "update")
		}
		fmt.Fprintf(w, `
			if !update.Dirty {
				return "", nil, errors.New("%s update failure, nothing to update")
			}`, entityName,
		)
		fmt.Fprint(w, `
			buf.WriteString(" SET ")
			buf.ReadFrom(update)
			buf.WriteString(" WHERE ")`)
		for i, c := range u.Columns {
			if i != 0 {
				fmt.Fprint(w, `
					update.WriteString(" AND ")`)
			}
			fmt.Fprintf(w, `
				update.WriteString(%s)
				update.WriteString("=")
				update.WritePlaceholder()
				update.Add(%s)`,
				g.Formatter.Identifier("table", table.Name, "column", c.Name),
				g.Formatter.IdentifierPrivate(columnForeignName(c)),
			)
		}
		fmt.Fprintf(w, `
			buf.ReadFrom(update)
			buf.WriteString(" RETURNING ")
			if len(r.%s) > 0 {
				buf.WriteString(strings.Join(r.%s, ", "))
			} else {`,
			g.Formatter.Identifier("columns"),
			g.Formatter.Identifier("columns"),
		)

		fmt.Fprint(w, `
		buf.WriteString("`)
		selectList(w, table, -1)
		fmt.Fprint(w, `")
	}`)
		fmt.Fprint(w, `
		return buf.String(), update.Args(), nil
	}`)
	}
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraint(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	for _, u := range g.uniqueConstraints(table) {
		method := []string{"updateOneBy"}
		arguments := ""
		arguments2 := ""

		for i, c := range u.Columns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
				arguments2 += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, ModeMandatory))
			arguments2 += g.Formatter.IdentifierPrivate(columnForeignName(c))
		}

		fmt.Fprintf(w, `
			func (r *%sRepositoryBase) %s(ctx context.Context, %s, p *%sPatch) (*%sEntity, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
			entityName,
		)

		fmt.Fprintf(w, `
			query, args, err := r.%s(%s, p)
			if err != nil {
				return nil, err
			}`,
			g.Formatter.Identifier(append(method, "query")...),
			arguments2,
		)
		fmt.Fprintf(w, `
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

		fmt.Fprintf(w, `
			if err != nil {
				if r.%s {
					r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.%s, "error", err.Error())
				}
				return nil, err
			}

			if r.%s {
				r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query success", "query", query, "table", r.%s)
			}

			return &ent, nil
		}`,
			g.Formatter.Identifier("debug"),
			g.Formatter.Identifier("log"),
			g.Formatter.Identifier("table"),
			g.Formatter.Identifier("debug"),
			g.Formatter.Identifier("log"),
			g.Formatter.Identifier("table"),
		)
	}
}

func (g *Generator) generateRepositoryUpsertQuery(w io.Writer, t *pqt.Table) {
	if g.Version < 9.5 {
		return
	}
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %sQuery(e *%sEntity, p *%sPatch, inf ...string) (string, []interface{}, error) {`,
		entityName,
		g.Formatter.Identifier("upsert"),
		entityName,
		entityName,
	)
	fmt.Fprintf(w, `
		upsert := NewComposer(%d)
		columns := bytes.NewBuffer(nil)
		buf := bytes.NewBufferString("INSERT INTO ")
		buf.WriteString(r.%s)
	`, len(t.Columns)*2, g.Formatter.Identifier("table"))

	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositoryInsertClause(w, c, "upsert")
	}

	fmt.Fprint(w, `
		if upsert.Dirty {
			buf.WriteString(" (")
			buf.ReadFrom(columns)
			buf.WriteString(") VALUES (")
			buf.ReadFrom(upsert)
			buf.WriteString(")")
		}
		buf.WriteString(" ON CONFLICT ")`,
	)

	fmt.Fprint(w, `
		if len(inf) > 0 {
		upsert.Dirty=false`)
	for _, c := range t.Columns {
		if c.IsDynamic {
			continue
		}
		g.generateRepositorySetClause(w, c, "upsert")
	}
	closeBrace(w, 1)

	fmt.Fprintf(w, `
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
	fmt.Fprint(w, `
		buf.WriteString("`)
	selectList(w, t, -1)
	fmt.Fprint(w, `")
	}`)
	fmt.Fprint(w, `
		}
		return buf.String(), upsert.Args(), nil
	}`)
}

func (g *Generator) generateRepositoryUpsert(w io.Writer, table *pqt.Table) {
	if g.Version < 9.5 {
		return
	}

	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {`,
		entityName,
		g.Formatter.Identifier("upsert"),
		entityName,
		entityName,
		entityName,
	)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(e, p, inf...)
			if err != nil {
				return nil, err
			}
			if err := r.%s.QueryRowContext(ctx, query, args...).Scan(`,
		g.Formatter.Identifier("upsert"),
		g.Formatter.Identifier("db"),
	)

	for _, c := range table.Columns {
		fmt.Fprintf(w, "&e.%s,\n", g.Formatter.Identifier(c.Name))
	}
	fmt.Fprintf(w, `); err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "upsert query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return nil, err
		}
		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "upsert query success", "query", query, "table", r.%s)
		}
		return e, nil
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositoryWhereClause(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		func %sCriteriaWhereClause(comp *Composer, c *%sCriteria, id int) (error) {`, g.Formatter.Identifier(t.Name), g.Formatter.Identifier(t.Name))
ColumnsLoop:
	for _, c := range t.Columns {
		braces := 0
		for _, plugin := range g.Plugins {
			if txt := plugin.WhereClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(txt)
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(w, map[string]interface{}{
					"selector": fmt.Sprintf("c.%s", g.Formatter.Identifier(c.Name)),
					"column":   g.sqlSelector(c, "id"),
					"composer": "comp",
					"id":       "id",
				}); err != nil {
					panic(err)
				}
				fmt.Fprintln(w, "")
				continue ColumnsLoop
			}
		}
		if g.columnType(c, ModeCriteria) == "<nil>" {
			return
		}
		if g.canBeNil(c, ModeCriteria) {
			braces++
			fmt.Fprintf(w, `
				if c.%s != nil {`, g.Formatter.Identifier(c.Name))
		}
		if g.isNullable(c, ModeCriteria) {
			braces++
			fmt.Fprintf(w, `
				if c.%s.Valid {`, g.Formatter.Identifier(c.Name))
		}
		if g.isType(c, ModeCriteria, "time.Time") {
			braces++
			fmt.Fprintf(w, `
				if !c.%s.IsZero() {`, g.Formatter.Identifier(c.Name))
		}

		fmt.Fprint(w,
			`if comp.Dirty {
				comp.WriteString(" AND ")
			}`)

		if c.IsDynamic {
			fmt.Fprintf(w, `
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
					fmt.Fprint(w, `
					if _, err := comp.WriteString(", "); err != nil {
						return err
					}`)
				}
				fmt.Fprintf(w, `
					if err := comp.WriteAlias(id); err != nil {
						return err
					}
					if _, err := comp.WriteString(%s); err != nil {
						return err
					}`,
					g.Formatter.Identifier("table", c.Columns[i].Table.Name, "column", c.Columns[i].Name),
				)
			}
			fmt.Fprint(w, `
				if _, err := comp.WriteString(")"); err != nil {
					return err
				}`)
		} else {
			fmt.Fprintf(w, `
				if err := comp.WriteAlias(id); err != nil {
					return err
				}
				if _, err := comp.WriteString(%s); err != nil {
					return err
				}`,
				g.Formatter.Identifier("table", t.Name, "column", c.Name),
			)
		}

		fmt.Fprintf(w, `
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
		closeBrace(w, braces)
		fmt.Fprintln(w, "")
	}
	fmt.Fprint(w, `
	return nil`)
	closeBrace(w, 1)
}

func (g *Generator) generateRepositoryJoinClause(w io.Writer, s *pqt.Schema) {
	fmt.Fprint(w, `
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

func selectList(w io.Writer, t *pqt.Table, nb int) {
	for i, c := range t.Columns {
		if i != 0 {
			fmt.Fprint(w, ", ")
		}
		if c.IsDynamic {
			fmt.Fprintf(w, "%s(", c.Func.Name)
			for i, arg := range c.Func.Args {
				if arg.Type != c.Columns[i].Type {
					fmt.Printf("wrong function (%s) argument type, expected %v but got %v\n", c.Func.Name, arg.Type, c.Columns[i].Type)
				}
				if i != 0 {
					fmt.Fprint(w, ", ")
				}
				if nb > -1 {
					fmt.Fprintf(w, "t%d.%s", nb, c.Columns[i].Name)
				} else {
					fmt.Fprintf(w, "%s", c.Columns[i].Name)
				}
			}
			fmt.Fprintf(w, ") AS %s", c.Name)
		} else {
			if nb > -1 {
				fmt.Fprintf(w, "t%d.%s", nb, c.Name)
			} else {
				fmt.Fprintf(w, "%s", c.Name)
			}
		}
	}
}

func (g *Generator) generateRepositoryFindQuery(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %sQuery(fe *%sFindExpr) (string, []interface{}, error) {`, entityName, g.Formatter.Identifier("find"), entityName)
	fmt.Fprintf(w, `
		comp := NewComposer(%d)
		buf := bytes.NewBufferString("SELECT ")
		if len(fe.%s) == 0 {
		buf.WriteString("`, len(t.Columns), g.Formatter.Identifier("columns"))
	selectList(w, t, 0)
	fmt.Fprintf(w, `")
		} else {
			buf.WriteString(strings.Join(fe.%s, ", "))
		}`, g.Formatter.Identifier("columns"))
	for nb, r := range joinableRelationships(t) {
		fmt.Fprintf(w, `
			if fe.%s != nil && fe.%s.%s {`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("fetch"),
		)
		fmt.Fprint(w, `
		buf.WriteString(", `)
		selectList(w, r.InversedTable, nb+1)
		fmt.Fprint(w, `")`)
		closeBrace(w, 1)
	}
	fmt.Fprintf(w, `
		buf.WriteString(" FROM ")
		buf.WriteString(r.%s)
		buf.WriteString(" AS t0")`, g.Formatter.Identifier("table"))
	for nb, r := range joinableRelationships(t) {
		oc := r.OwnerColumns
		ic := r.InversedColumns
		if len(oc) != len(ic) {
			panic("number of owned and inversed foreign key columns is not equal")
		}

		fmt.Fprintf(w, `
			if fe.%s != nil {`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
		)
		fmt.Fprintf(w, `
			joinClause(comp, fe.%s.%s, "%s AS t%d ON `,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("kind"),
			r.InversedTable.FullName(),
			nb+1,
		)

		for i := 0; i < len(oc); i++ {
			if i > 0 {
				fmt.Fprint(w, `AND `)
			}
			fmt.Fprintf(w, `t%d.%s=t%d.%s`, 0, oc[i].Name, nb+1, ic[i].Name)
		}
		fmt.Fprint(w, `")`)

		fmt.Fprintf(w, `
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

		closeBrace(w, 1)
	}

	fmt.Fprintf(w, `
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
		fmt.Fprintf(w, `
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

	fmt.Fprint(w, `
		if comp.Dirty {
			//fmt.Println("comp", comp.String())
			//fmt.Println("buf", buf.String())
			if _, err := buf.WriteString(" WHERE "); err != nil {
				return "", nil, err
			}
			buf.ReadFrom(comp)
			//fmt.Println("comp - after", comp.String())
			//fmt.Println("buf - after", buf.String())
		}
	`)

	fmt.Fprintf(w, `
	if len(fe.%s) > 0 {
		i:=0
		comp.WriteString(" ORDER BY ")

		for cn, asc := range fe.%s {
			for _, tcn := range %s {
				if cn == tcn {
					if i > 0 {
						if _, err := comp.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if _, err := comp.WriteString(cn); err != nil {
						return "", nil, err
					}
					if !asc {
						if _, err := comp.WriteString(" DESC "); err != nil {
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

	fmt.Fprint(w, `
		buf.ReadFrom(comp)
	`)
	fmt.Fprintln(w, `
	return buf.String(), comp.Args(), nil
}`)
}

func (g *Generator) generateRepositoryFind(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) ([]*%sEntity, error) {`, entityName, g.Formatter.Identifier("find"), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("db"),
	)

	fmt.Fprintf(w, `
		if err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "find query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return nil, err
		}
			defer rows.Close()

		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "find query success", "query", query, "table", r.%s)
		}

`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)

	fmt.Fprintf(w, `
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
		fmt.Fprint(w, `
		var prop []interface{}`)
	}
	for _, r := range joinableRelationships(t) {
		if r.Type == pqt.RelationshipTypeOneToMany || r.Type == pqt.RelationshipTypeManyToMany {
			continue
		}
		fmt.Fprintf(w, `
			if fe.%s != nil && fe.%s.%s {
				ent.%s = &%sEntity{}
				if prop, err = ent.%s.%s(); err != nil {
					return nil, err
				}
				props = append(props, prop...)
			}`,
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("fetch"),
			g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier(r.InversedTable.Name),
			g.Formatter.Identifier(or(r.InversedName, r.InversedTable.Name)),
			g.Formatter.Identifier("props"),
		)
	}
	fmt.Fprint(w, `
			err = rows.Scan(props...)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &ent)
		}`)
	fmt.Fprintf(w, `
		if err = rows.Err(); err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return nil, err
		}
		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "find query success", "query", query, "table", r.%s)
		}
		return entities, nil
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositoryFindIter(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, fe *%sFindExpr) (*%sIterator, error) {`, entityName, g.Formatter.Identifier("findIter"), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(fe)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}
			return &%sIterator{
				rows: rows,
				cols: []string{`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("db"),
		g.Formatter.Identifier(t.Name),
	)
	for _, c := range t.Columns {
		fmt.Fprintf(w, `"%s",`, c.Name)
	}
	fmt.Fprint(w, `},
		}, nil
	}`)
}

func (g *Generator) generateRepositoryCount(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCountExpr) (int64, error) {`, entityName, g.Formatter.Identifier("count"), entityName)
	fmt.Fprintf(w, `
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
		fmt.Fprintf(w, `
		%s: c.%s,`, g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)), g.Formatter.Identifier("join", or(r.InversedName, r.InversedTable.Name)))
	}
	fmt.Fprintf(w, `
		})
		if err != nil {
			return 0, err
		}
		var count int64
		if err := r.%s.QueryRowContext(ctx, query, args...).Scan(&count)`,
		g.Formatter.Identifier("db"),
	)

	fmt.Fprintf(w, `; err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "count query failure", "query", query, "table", r.%s, "error", err.Error())
			}
			return 0, err
		}

		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "count query success", "query", query, "table", r.%s)
		}

		return count, nil
	}`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (*%sEntity, error) {`,
		entityName,
		g.Formatter.Identifier("FindOneBy", pk.Name),
		g.columnType(pk, ModeMandatory),
		entityName,
	)
	fmt.Fprintf(w, `
		find := NewComposer(%d)
		find.WriteString("SELECT ")
		if len(r.%s) == 0 {
			find.WriteString("`,
		len(table.Columns), g.Formatter.Identifier("columns"))
	selectList(w, table, -1)
	fmt.Fprintf(w, `")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, g.Formatter.Identifier("columns"))

	fmt.Fprintf(w, `
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
		g.Formatter.Identifier("table", table.Name),
		g.Formatter.Identifier("table", table.Name, "column", pk.Name),
		entityName,
	)

	fmt.Fprintf(w, `
		props, err := ent.%s(r.%s...)
		if err != nil {
			return nil, err
		}
		err = r.%s.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)`,
		g.Formatter.Identifier("props"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
	)
	fmt.Fprintf(w, `
		if err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "find by primary key query failure", "query", find.String(), "table", r.%s, "error", err.Error())
			}
			return nil, err
		}

		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "find by primary key query success", "query", find.String(), "table", r.%s)
		}

		return &ent, nil
	}`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("table"),
	)
}

func (g *Generator) generateRepositoryFindOneByUniqueConstraint(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	for _, u := range g.uniqueConstraints(table) {
		method := []string{"FindOneBy"}
		arguments := ""

		for i, c := range u.Columns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, ModeMandatory))
		}

		fmt.Fprintf(w, `
			func (r *%sRepositoryBase) %s(ctx context.Context, %s) (*%sEntity, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
		)
		fmt.Fprintf(w, `
			find := NewComposer(%d)
			find.WriteString("SELECT ")
					if len(r.%s) == 0 {
			find.WriteString("`,
			len(table.Columns), g.Formatter.Identifier("columns"))
		selectList(w, table, -1)
		fmt.Fprintf(w, `")
		} else {
			find.WriteString(strings.Join(r.%s, ", "))
		}`, g.Formatter.Identifier("columns"))

		fmt.Fprintf(w, `
			find.WriteString(" FROM ")
			find.WriteString(%s)
			find.WriteString(" WHERE ")`,
			g.Formatter.Identifier("table", table.Name),
		)
		for i, c := range u.Columns {
			if i != 0 {
				fmt.Fprint(w, `find.WriteString(" AND ")`)
			}
			fmt.Fprintf(w, `
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(%s)
		`, g.Formatter.Identifier("table", table.Name, "column", c.Name), g.Formatter.IdentifierPrivate(columnForeignName(c)))
		}

		fmt.Fprintf(w, `
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
		fmt.Fprint(w, `
			if err != nil {
				return nil, err
			}

			return &ent, nil
		}`)
	}
}

func (g *Generator) generateRepositoryDeleteOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (int64, error) {`,
		entityName,
		g.Formatter.Identifier("DeleteOneBy", pk.Name),
		g.columnType(pk, ModeMandatory),
	)
	fmt.Fprintf(w, `
		find := NewComposer(%d)
		find.WriteString("DELETE FROM ")
		find.WriteString(%s)
		find.WriteString(" WHERE ")
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(pk)`, len(table.Columns),
		g.Formatter.Identifier("table", table.Name),
		g.Formatter.Identifier("table", table.Name, "column", pk.Name),
	)

	fmt.Fprintf(w, `
		res, err := r.%s.ExecContext(ctx, find.String(), find.Args()...)`,
		g.Formatter.Identifier("db"),
	)
	fmt.Fprint(w, `
		if err != nil {
				return 0, err
			}

		return res.RowsAffected()
	}`)
}

func (g *Generator) generateEntityProp(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		func (e *%sEntity) %s(cn string) (interface{}, bool) {`, g.Formatter.Identifier(t.Name), g.Formatter.Identifier("prop"))
	fmt.Fprintln(w, `
		switch cn {`)

ColumnsLoop:
	for _, c := range t.Columns {
		fmt.Fprintf(w, `
			case %s:`, g.Formatter.Identifier("table", t.Name, "column", c.Name))
		for _, plugin := range g.Plugins {
			if txt := plugin.ScanClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(fmt.Sprintf(`
					return %s, true`, txt))
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(w, map[string]interface{}{
					"selector": fmt.Sprintf("e.%s", g.Formatter.Identifier(c.Name)),
				}); err != nil {
					panic(err)
				}
				fmt.Fprintln(w, "")
				continue ColumnsLoop
			}
		}
		switch {
		case g.isArray(c, ModeDefault):
			pn := g.Formatter.Identifier(c.Name)
			switch g.columnType(c, ModeDefault) {
			case "pq.Int64Array":
				fmt.Fprintf(w, `if e.%s == nil { e.%s = []int64{} }`, pn, pn)
			case "pq.StringArray":
				fmt.Fprintf(w, `if e.%s == nil { e.%s = []string{} }`, pn, pn)
			case "pq.Float64Array":
				fmt.Fprintf(w, `if e.%s == nil { e.%s = []float64{} }`, pn, pn)
			case "pq.BoolArray":
				fmt.Fprintf(w, `if e.%s == nil { e.%s = []bool{} }`, pn, pn)
			case "pq.ByteaArray":
				fmt.Fprintf(w, `if e.%s == nil { e.%s = [][]byte{} }`, pn, pn)
			}

			fmt.Fprintf(w, `
				return &e.%s, true`, g.Formatter.Identifier(c.Name))
		case g.canBeNil(c, ModeDefault):
			fmt.Fprintf(w, `
				return e.%s, true`,
				g.Formatter.Identifier(c.Name),
			)
		default:
			fmt.Fprintf(w, `
				return &e.%s, true`,
				g.Formatter.Identifier(c.Name),
			)
		}
	}

	fmt.Fprint(w, `
		default:`)
	fmt.Fprint(w, `
		return nil, false`)
	fmt.Fprint(w, "}\n}\n")

}

func (g *Generator) generateEntityProps(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, `
		func (e *%sEntity) %s(cns ...string) ([]interface{}, error) {`, g.Formatter.Identifier(t.Name), g.Formatter.Identifier("props"))
	fmt.Fprintf(w, `
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
	fmt.Fprint(w, `
		}`)
}

func (g *Generator) generateRepositoryScanRows(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	fmt.Fprintf(w, `
		func %s(rows *sql.Rows) (entities []*%sEntity, err error) {`, g.Formatter.Identifier("scan", t.Name, "rows"), entityName)
	fmt.Fprintf(w, `
		for rows.Next() {
			var ent %sEntity
			err = rows.Scan(`,
		entityName,
	)
	for _, c := range t.Columns {
		fmt.Fprintf(w, "&ent.%s,\n", g.Formatter.Identifier(c.Name))
	}
	fmt.Fprint(w, `)
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
			if ct, ok := mapto.(CustomType); ok {
				switch m {
				case ModeMandatory:
					return ct.mandatoryTypeOf.Kind() == reflect.Ptr || ct.mandatoryTypeOf.Kind() == reflect.Map
				case ModeOptional:
					return ct.optionalTypeOf.Kind() == reflect.Ptr || ct.optionalTypeOf.Kind() == reflect.Map
				case ModeCriteria:
					return ct.criteriaTypeOf.Kind() == reflect.Ptr || ct.criteriaTypeOf.Kind() == reflect.Map
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
			if ct, ok := mapto.(CustomType); ok {
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

func (g *Generator) generateStatics(w io.Writer, s *pqt.Schema) {
	fmt.Fprint(w, `

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
		return fmt.Errorf("expected to get source argument in format '[text1,text2,...,textN]', but got %s", srcs)
	}

	if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
		return fmt.Errorf("expected to get source argument in format '[text1,text2,...,textN]', but got %s", srcs)
	}

	*a = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)

	return nil
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayString) Value() (driver.Value, error) {
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
		if _, err = buffer.WriteString(v); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
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
	PlaceholderFunc, SelectorFunc string
	Cast                          string
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
			fmt.Fprint(w, txt)
			fmt.Fprint(w, "\n\n")
		}
	}
}

func (g *Generator) uniqueConstraints(t *pqt.Table) []*pqt.Constraint {
	var unique []*pqt.Constraint
	for _, c := range tableConstraints(t) {
		if c.Type == pqt.ConstraintTypeUnique {
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
