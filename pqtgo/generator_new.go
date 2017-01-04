package pqtgo

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"reflect"
	"strings"

	"text/template"

	"github.com/piotrkowalczuk/pqt"
)

type Context struct {
	Table     *pqt.Table
	Column    *pqt.Column
	Schema    *pqt.Schema
	Selectors map[string]interface{}
	Formatter *Formatter
}

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

type Gen struct {
	Formatter *Formatter
	Pkg       string
	Imports   []string
	Plugins   []Plugin
}

// Generate ...
func (g *Gen) Generate(s *pqt.Schema) ([]byte, error) {
	code, err := g.generate(s)
	if err != nil {
		return nil, err
	}

	return code.Bytes(), nil
}

// GenerateTo ...
func (g *Gen) GenerateTo(w io.Writer, s *pqt.Schema) error {
	code, err := g.generate(s)
	if err != nil {
		return err
	}

	_, err = code.WriteTo(w)
	return err
}

func (g *Gen) generatePackage(w io.Writer) {
	pkg := g.Pkg
	if pkg == "" {
		pkg = "main"
	}
	fmt.Fprintf(w, "package %s\n", pkg)
}

func (g *Gen) generateImports(w io.Writer, schema *pqt.Schema) {
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

func (g *Gen) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	b := bytes.NewBuffer(nil)

	g.generatePackage(b)
	g.generateImports(b, s)
	for _, t := range s.Tables {
		g.generateConstants(b, t)
		g.generateColumns(b, t)
		g.generateEntity(b, t)
		g.generateEntityProp(b, t)
		g.generateEntityProps(b, t)
		g.generateIterator(b, t)
		g.generateCriteria(b, t)
		g.generatePatch(b, t)
		g.generateRepositoryScanRows(b, t)
		g.generateRepository(b, t)
		g.generateRepositoryInsertQuery(b, t)
		g.generateRepositoryInsert(b, t)
		g.generateRepositoryFindQuery(b, t)
		g.generateRepositoryFind(b, t)
		g.generateRepositoryFindIter(b, t)
		g.generateRepositoryFindOneByPrimaryKey(b, t)
		g.generateRepositoryFindOneByUniqueConstraint(b, t)
		g.generateRepositoryUpdateOneByPrimaryKeyQuery(b, t)
		g.generateRepositoryUpdateOneByPrimaryKey(b, t)
		g.generateRepositoryCount(b, t)
		g.generateRepositoryDeleteOneByPrimaryKey(b, t)
	}
	g.generateStatics(b)

	return b, nil
}

func (g *Gen) generateEntity(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "// %sEntity ...\n", g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, "type %sEntity struct{\n", g.Formatter.Identifier(t.Name))
	for prop := range g.entityPropertiesGenerator(t) {
		fmt.Fprintf(w, "// %s ...\n", g.Formatter.Identifier(prop.Name))
		if prop.Tags != "" {
			fmt.Fprintf(w, "%s %s %s\n", g.Formatter.Identifier(prop.Name), prop.Type, prop.Tags)
		} else {
			fmt.Fprintf(w, "%s %s\n", g.Formatter.Identifier(prop.Name), prop.Type)
		}
	}
	fmt.Fprint(w, "}\n\n")
}

func (g *Gen) generateCriteria(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "type %sCriteria struct {\n", g.Formatter.Identifier(t.Name))
	fmt.Fprintf(w, "%s, %s int64\n", g.Formatter.Identifier("offset"), g.Formatter.Identifier("limit"))
	fmt.Fprintf(w, "%s map[string]bool\n", g.Formatter.Identifier("sort"))

	for _, c := range t.Columns {
		if t := g.columnType(c, modeCriteria); t != "<nil>" {
			fmt.Fprintf(w, "%s %s\n", g.Formatter.Identifier(c.Name), t)
		}
	}
	fmt.Fprintln(w, "}")
}

func (g *Gen) generatePatch(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "type %sPatch struct {\n", g.Formatter.Identifier(t.Name))

ArgumentsLoop:
	for _, c := range t.Columns {
		if c.PrimaryKey {
			continue ArgumentsLoop
		}

		if t := g.columnType(c, modeOptional); t != "<nil>" {
			fmt.Fprintf(w, "%s %s\n", g.Formatter.Identifier(c.Name), t)
		}
	}
	fmt.Fprintln(w, "}")
}

func (g *Gen) generateIterator(w io.Writer, t *pqt.Table) {
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

// Columns is wrapper around sql.Rows.Columns method, that also cache outpu inside iterator.
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
	cols, err := i.rows.Columns()
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
func (g *Gen) entityPropertiesGenerator(t *pqt.Table) chan structField {
	fields := make(chan structField)

	go func(out chan structField) {
		for _, c := range t.Columns {
			if t := g.columnType(c, modeDefault); t != "<nil>" {
				out <- structField{Name: g.Formatter.Identifier(c.Name), Type: t}
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

func (g *Gen) generateRepository(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `type %sRepositoryBase struct {
			%s string
			%s []string
			%s *sql.DB
			%s bool
			%s log.Logger
		}
`,
		g.Formatter.Identifier(table.Name),
		g.Formatter.Identifier("table"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
	)
}

func (g *Gen) generateColumns(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, "var (\n %s  = []string{\n", g.Formatter.Identifier("table", table.Name, "columns"))

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(w, "%s", g.Formatter.Identifier("table", table.Name, "column", name))
		io.WriteString(w, ",")
		io.WriteString(w, "\n")
	}
	io.WriteString(w, "})\n")
}

func (g *Gen) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	g.generateConstantsColumns(code, table)
	g.generateConstantsConstraints(code, table)
	code.WriteString(")\n")
}

func (g *Gen) generateConstantsColumns(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `%s = "%s"
	`, g.Formatter.Identifier("table", table.Name), table.FullName())

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(w, `%s = "%s"
		`, g.Formatter.Identifier("table", table.Name, "column", name), name)
	}
}

func (g *Gen) generateConstantsConstraints(w io.Writer, table *pqt.Table) {
	for _, c := range tableConstraints(table) {
		name := fmt.Sprintf("%s", pqt.JoinColumns(c.Columns, "_"))
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraint", name, "Check"), c.String())
		case pqt.ConstraintTypePrimaryKey:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "constraintPrimaryKey"), c.String())
		case pqt.ConstraintTypeForeignKey:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "Constraint", name, "ForeignKey"), c.String())
		case pqt.ConstraintTypeExclusion:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "Constraint", name, "Exclusion"), c.String())
		case pqt.ConstraintTypeUnique:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "Constraint", name, "Unique"), c.String())
		case pqt.ConstraintTypeIndex:
			fmt.Fprintf(w, `%s = "%s"`, g.Formatter.Identifier("table", table.Name, "Constraint", name, "Index"), c.String())
		}

		io.WriteString(w, "\n")
	}
}

func (g *Gen) generateRepositoryInsertQuery(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %sQuery(e *%sEntity) (string, []interface{}, error) {`, entityName, g.Formatter.Identifier("insert"), entityName)
	fmt.Fprintf(w, `
		ins := pqtgo.NewComposer(%d)
		buf := bytes.NewBufferString("INSERT INTO " + r.%s)
		col := bytes.NewBuffer(nil)
	`, len(table.Columns), g.Formatter.Identifier("table"))

ColumnsLoop:
	for _, c := range table.Columns {
		braces := 0

		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue ColumnsLoop
		default:
			if g.canBeNil(c, modeDefault) {
				fmt.Fprintf(w, "if e.%s != nil {\n", g.Formatter.Identifier(c.Name))
				braces++
			}
			if g.isNullable(c, modeDefault) {
				fmt.Fprintf(w, "if e.%s.Valid {\n", g.Formatter.Identifier(c.Name))
				braces++
			}
			if g.isType(c, modeDefault, "time.Time") {
				fmt.Fprintf(w, "if !e.%s.IsZero() {\n", g.Formatter.Identifier(c.Name))
				braces++
			}
			fmt.Fprintf(w,
				`if col.Len() > 0 {
						if _, err := col.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if _, err := col.WriteString(%s); err != nil {
						return "", nil, err
					}
					if ins.Dirty {
						if _, err := ins.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if err := ins.WritePlaceholder(); err != nil {
						return "", nil, err
					}
					ins.Add(e.%s)
					ins.Dirty=true`,
				g.Formatter.Identifier("table", table.Name, "column", c.Name),
				g.Formatter.Identifier(c.Name),
			)

			closeBrace(w, braces)
			fmt.Fprintln(w, "")
		}
	}
	fmt.Fprintf(w, `if col.Len() > 0 {
				buf.WriteString(" (")
				buf.ReadFrom(col)
				buf.WriteString(") VALUES (")
				buf.ReadFrom(ins)
				buf.WriteString(") ")
				if len(r.%s) > 0 {
					buf.WriteString("RETURNING ")
					buf.WriteString(strings.Join(r.%s, ", "))
				}
			}
			return buf.String(), ins.Args(), nil
}
`,
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("columns"),
	)
}

func (g *Gen) generateRepositoryInsert(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, e *%sEntity) (*%sEntity, error) {`, entityName, g.Formatter.Identifier("insert"), entityName, entityName)
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
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.Table, "error", err.Error())
			}
			return nil, err
		}
		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query success", "query", query, "table", r.Table)
		}
		return e, nil
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
	)
}

func (g *Gen) generateRepositoryUpdateOneByPrimaryKeyQuery(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, "func (r *%sRepositoryBase) %sQuery(pk %s, p *%sPatch) (string, []interface{}, error) {\n", entityName, g.Formatter.Identifier("UpdateOneBy", pk.Name), g.columnType(pk, modeMandatory), entityName)
	fmt.Fprintf(w, `buf := bytes.NewBufferString("UPDATE ")
		buf.WriteString(r.%s)
		update := pqtgo.NewComposer(%d)
	`, g.Formatter.Identifier("table"), len(table.Columns))
	fmt.Fprintln(w, "update.Add(pk)")

ColumnsLoop:
	for _, c := range table.Columns {
		if c == pk {
			continue ColumnsLoop
		}
		if g.canBeNil(c, modeOptional) {
			fmt.Fprintf(w, "if p.%s != nil {\n", g.Formatter.Identifier(c.Name))
		}
		if g.isNullable(c, modeOptional) {
			fmt.Fprintf(w, "if p.%s.Valid {\n", g.Formatter.Identifier(c.Name))
		}
		if g.isType(c, modeOptional, "time.Time") {
			fmt.Fprintf(w, "if !p.%s.IsZero() {", g.Formatter.Identifier(c.Name))
		}

		fmt.Fprintf(w, `if update.Dirty {
				if _, err := update.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := update.WriteString(%s); err != nil {
				return "", nil, err
			}
			if _, err := update.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := update.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			update.Add(p.%s)
			update.Dirty=true
		`, g.Formatter.Identifier("table", table.Name, "column", c.Name), g.Formatter.Identifier(c.Name))

		if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprintf(w, `
					} else {
						if update.Dirty {
							if _, err := update.WriteString(", "); err != nil {
								return "", nil, err
							}
						}
						if _, err := update.WriteString(%s); err != nil {
							return "", nil, err
						}
						if _, err := update.WriteString("=%s"); err != nil {
							return "", nil, err
						}
				`, g.Formatter.Identifier("table", table.Name, "column", c.Name), d)
			}
		}

		if g.isType(c, modeOptional, "time.Time") {
			fmt.Fprint(w, "\n}\n")
		}
		if g.canBeNil(c, modeOptional) {
			fmt.Fprint(w, "\n}\n")
		}
		if g.isNullable(c, modeOptional) {
			fmt.Fprint(w, "\n}\n")
		}
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
	buf.WriteString(strings.Join(r.%s, ", "))

	return buf.String(), update.Args(), nil
}
`, g.Formatter.Identifier("table", table.Name, "column", pk.Name), g.Formatter.Identifier("columns"))
}

func (g *Gen) generateRepositoryUpdateOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, pk %s, p *%sPatch) (*%sEntity, error) {`, entityName, g.Formatter.Identifier("updateOneBy", pk.Name), g.columnType(pk, modeMandatory), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(pk, p)
			if err != nil {
				return nil, err
			}
`, g.Formatter.Identifier("updateOneBy", pk.Name))

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
	fmt.Fprint(w, `
		if err != nil {
			return nil, err
		}

		return &ent, nil
}
`)
}

func (g *Gen) generateRepositoryFindQuery(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %sQuery(s []string, c *%sCriteria) (string, []interface{}, error) {`, entityName, g.Formatter.Identifier("find"), entityName)
	fmt.Fprintf(w, `
	where := pqtgo.NewComposer(%d)
	buf := bytes.NewBufferString("SELECT ")
	buf.WriteString(strings.Join(s, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(r.%s)
	buf.WriteString(" ")
`,
		len(table.Columns),
		g.Formatter.Identifier("table"),
	)

ColumnsLoop:
	for _, c := range table.Columns {
		for _, plugin := range g.Plugins {
			if txt := plugin.WhereClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(txt)
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(w, map[string]interface{}{
					"selector": fmt.Sprintf("c.%s", g.Formatter.Identifier(c.Name)),
					"composer": "where",
				}); err != nil {
					panic(err)
				}
					continue ColumnsLoop
			}
		}
		if g.canBeNil(c, modeOptional) {
			fmt.Fprintf(w, "if c.%s != nil {\n", g.Formatter.Identifier(c.Name))
		}
		if g.isNullable(c, modeOptional) {
			fmt.Fprintf(w, "if c.%s.Valid {\n", g.Formatter.Identifier(c.Name))
		}
		if g.isType(c, modeOptional, "time.Time") {
			fmt.Fprintf(w, "if !c.%s.IsZero() {", g.Formatter.Identifier(c.Name))
		}
		fmt.Fprintf(w,
			`if where.Dirty {
					where.WriteString(" AND ")
				}
				if _, err := where.WriteString(%s); err != nil {
					return "", nil, err
				}
				if _, err := where.WriteString("="); err != nil {
					return "", nil, err
				}
				if err := where.WritePlaceholder(); err != nil {
					return "", nil, err
				}
				where.Add(c.%s)
				where.Dirty=true`,
			g.Formatter.Identifier("table", table.Name, "column", c.Name),
			g.Formatter.Identifier(c.Name),
		)
		if g.isType(c, modeOptional, "time.Time") {
			fmt.Fprintln(w, "}")
		}
		if g.canBeNil(c, modeOptional) {
			fmt.Fprintln(w, "}")
		}
		if g.isNullable(c, modeOptional) {
			fmt.Fprintln(w, "}")
		}
		fmt.Fprintln(w, "")
	}

	fmt.Fprintln(w, `
	if where.Dirty {
		buf.WriteString("WHERE ")
		buf.ReadFrom(where)
	}
	return buf.String(), where.Args(), nil
}`)
}

func (g *Gen) generateRepositoryFind(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCriteria) ([]*%sEntity, error) {`, entityName, g.Formatter.Identifier("find"), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(r.%s, c)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
	)

	fmt.Fprintf(w, `
		if err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.Table, "error", err.Error())
			}
			return nil, err
		}
			defer rows.Close()

		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query success", "query", query, "table", r.Table)
		}

		return %s(rows)
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("Scan", table.Name, "rows"),
	)
}

func (g *Gen) generateRepositoryFindIter(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCriteria) (*%sIterator, error) {`, entityName, g.Formatter.Identifier("findIter"), entityName, entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery(r.%s, c)
			if err != nil {
				return nil, err
			}
			rows, err := r.%s.QueryContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}
			return &%sIterator{rows: rows}, nil
}
`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("columns"),
		g.Formatter.Identifier("db"),
		g.Formatter.Identifier(t.Name),
	)
}

func (g *Gen) generateRepositoryCount(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCriteria) (int64, error) {`, entityName, g.Formatter.Identifier("count"), entityName)
	fmt.Fprintf(w, `
			query, args, err := r.%sQuery([]string{"COUNT(*)"}, c)
			if err != nil {
				return 0, err
			}
			var count int64
			if err := r.%s.QueryRowContext(ctx, query, args...).Scan(&count)`,
		g.Formatter.Identifier("find"),
		g.Formatter.Identifier("db"),
	)

	fmt.Fprintf(w, `; err != nil {
			if r.%s {
				r.%s.Log("level", "error", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query failure", "query", query, "table", r.Table, "error", err.Error())
			}
			return 0, err
		}

		if r.%s {
			r.%s.Log("level", "debug", "timestamp", time.Now().Format(time.RFC3339), "msg", "insert query success", "query", query, "table", r.Table)
		}

		return count, nil
	}
`,
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
		g.Formatter.Identifier("debug"),
		g.Formatter.Identifier("log"),
	)
}

func (g *Gen) generateRepositoryFindOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (*%sEntity, error) {`,
		entityName,
		g.Formatter.Identifier("FindOneBy", pk.Name),
		g.columnType(pk, modeMandatory),
		entityName,
	)
	fmt.Fprintf(w, `
	find := pqtgo.NewComposer(%d)
	find.WriteString("SELECT ")
	find.WriteString(strings.Join(r.%s, ", "))
	find.WriteString(" FROM ")
	find.WriteString(%s)
	find.WriteString(" WHERE ")
	find.WriteString(%s)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(pk)
	var (
		ent %sEntity
	)
`, len(table.Columns),
		g.Formatter.Identifier("columns"),
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
	fmt.Fprint(w, `
		if err != nil {
			return nil, err
		}

		return &ent, nil
}
`)
}

func (g *Gen) generateRepositoryFindOneByUniqueConstraint(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	var unique []*pqt.Constraint
	for _, c := range tableConstraints(table) {
		if c.Type == pqt.ConstraintTypeUnique {
			unique = append(unique, c)
		}
	}
	if len(unique) < 1 {
		return
	}

	for _, u := range unique {
		method := []string{"FindOneBy"}
		arguments := ""

		for i, c := range u.Columns {
			if i != 0 {
				method = append(method, "And")
				arguments += ", "
			}
			method = append(method, c.Name)
			arguments += fmt.Sprintf("%s %s", g.Formatter.IdentifierPrivate(columnForeignName(c)), g.columnType(c, modeMandatory))
		}

		fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, %s) (*%sEntity, error) {`,
			entityName,
			g.Formatter.Identifier(method...),
			arguments,
			entityName,
		)
		fmt.Fprintf(w, `
	find := pqtgo.NewComposer(%d)
	find.WriteString("SELECT ")
	find.WriteString(strings.Join(r.%s, ", "))
	find.WriteString(" FROM ")
	find.WriteString(%s)
	find.WriteString(" WHERE ")`,
			len(table.Columns),
			g.Formatter.Identifier("columns"),
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
}
`)
	}
}

func (g *Gen) generateRepositoryDeleteOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.Formatter.Identifier(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (int64, error) {`,
		entityName,
		g.Formatter.Identifier("DeleteOneBy", pk.Name),
		g.columnType(pk, modeMandatory),
	)
	fmt.Fprintf(w, `
	find := pqtgo.NewComposer(%d)
	find.WriteString("DELETE FROM ")
	find.WriteString(%s)
	find.WriteString(" WHERE ")
	find.WriteString(%s)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(pk)
`, len(table.Columns),
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
}
`)
}

func (g *Gen) generateEntityProp(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "func (e *%sEntity) %s(cn string) (interface{}, bool) {\n", g.Formatter.Identifier(t.Name), g.Formatter.Identifier("prop"))
	fmt.Fprintln(w, "switch cn {")
	for _, c := range t.Columns {
		fmt.Fprintf(w, "case %s:\n", g.Formatter.Identifier("table", t.Name, "column", c.Name))
		if g.canBeNil(c, modeDefault) && !g.isArray(c, modeDefault) {
			fmt.Fprintf(w, "return e.%s, true\n", g.Formatter.Identifier(c.Name))
		} else {
			fmt.Fprintf(w, "return &e.%s, true\n", g.Formatter.Identifier(c.Name))
		}
	}
	fmt.Fprint(w, "default:\n")
	fmt.Fprint(w, "return nil, false\n")
	fmt.Fprint(w, "}\n}\n")
}

func (g *Gen) generateEntityProps(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "func (e *%sEntity) %s(cns ...string) ([]interface{}, error) {\n", g.Formatter.Identifier(t.Name), g.Formatter.Identifier("props"))
	fmt.Fprintf(w, `res := make([]interface{}, 0, len(cns))
		for _, cn := range cns {
			if prop, ok := e.%s(cn); ok {
				res = append(res, prop)
			} else {
				return nil, fmt.Errorf("unexpected column provided: %%s", cn)
			}
		}
		return res, nil`, g.Formatter.Identifier("prop"))
	fmt.Fprint(w, "\n}\n")
}

func (g *Gen) generateRepositoryScanRows(w io.Writer, t *pqt.Table) {
	entityName := g.Formatter.Identifier(t.Name)
	fmt.Fprintf(w, `func %s(rows *sql.Rows) (entities []*%sEntity, err error) {
	`, g.Formatter.Identifier("scan", t.Name, "rows"), entityName)
	fmt.Fprintf(w, `for rows.Next() {
		var ent %sEntity
		err = rows.Scan(
	`, entityName)
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
}
`)
}

func (g *Gen) isArray(c *pqt.Column, m int32) bool {
	return strings.HasPrefix(g.columnType(c, m), "[]")
}

func (g *Gen) canBeNil(c *pqt.Column, m int32) bool {
	if tp, ok := c.Type.(pqt.MappableType); ok {
		for _, mapto := range tp.Mapping {
			if ct, ok := mapto.(CustomType); ok {
				switch m {
				case modeMandatory:
					return ct.mandatoryTypeOf.Kind() == reflect.Ptr || ct.mandatoryTypeOf.Kind() == reflect.Map
				case modeOptional:
					return ct.optionalTypeOf.Kind() == reflect.Ptr || ct.optionalTypeOf.Kind() == reflect.Map
				case modeCriteria:
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

func (g *Gen) columnType(c *pqt.Column, m int32) string {
	switch m {
	case modeCriteria:
	case modeMandatory:
	case modeOptional:
	default:
		if c.NotNull || c.PrimaryKey {
			m = modeMandatory
		}
	}
	for _, plugin := range g.Plugins {
		if txt := plugin.PropertyType(c, m); txt != "" {
			return txt
		}
	}
	return g.Formatter.Type(c.Type, m)
}

func (g *Gen) isType(c *pqt.Column, m int32, types ...string) bool {
	for _, t := range types {
		if g.columnType(c, m) == t {
			return true
		}
	}
	return false
}

func (g *Gen) isNullable(c *pqt.Column, m int32) bool {
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

func (g *Gen) generateStatics(w io.Writer) {
	fmt.Fprint(w, `
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
`)
}

func generateTypeBuiltin(t BuiltinType, m int32) (r string) {
	switch types.BasicKind(t) {
	case types.Bool:
		r = chooseType("bool", "*bool", "*bool", m)
	case types.Int:
		r = chooseType("int", "*int", "*int", m)
	case types.Int8:
		r = chooseType("int8", "*int8", "*int8", m)
	case types.Int16:
		r = chooseType("int16", "*int16", "*int16", m)
	case types.Int32:
		r = chooseType("int32", "*int32", "*int32", m)
	case types.Int64:
		r = chooseType("int64", "*int64", "*int64", m)
	case types.Uint:
		r = chooseType("uint", "*uint", "*uint", m)
	case types.Uint8:
		r = chooseType("uint8", "*uint8", "*uint8", m)
	case types.Uint16:
		r = chooseType("uint16", "*uint16", "*uint16", m)
	case types.Uint32:
		r = chooseType("uint32", "*uint32", "*uint32", m)
	case types.Uint64:
		r = chooseType("uint64", "*uint64", "*uint64", m)
	case types.Float32:
		r = chooseType("float32", "*float32", "*float32", m)
	case types.Float64:
		r = chooseType("float64", "*float64", "*float64", m)
	case types.Complex64:
		r = chooseType("complex64", "*complex64", "*complex64", m)
	case types.Complex128:
		r = chooseType("complex128", "*complex128", "*complex128", m)
	case types.String:
		r = chooseType("string", "*string", "*string", m)
	default:
		r = "invalid"
	}

	return
}

func generateTypeBase(t pqt.Type, m int32) string {
	switch t {
	case pqt.TypeText():
		return chooseType("string", "sql.NullString", "sql.NullString", m)
	case pqt.TypeBool():
		return chooseType("bool", "sql.NullBool", "sql.NullBool", m)
	case pqt.TypeIntegerSmall():
		return chooseType("int16", "*int16", "*int16", m)
	case pqt.TypeInteger():
		return chooseType("int32", "*int32", "*int32", m)
	case pqt.TypeIntegerBig():
		return chooseType("int64", "sql.NullInt64", "sql.NullInt64", m)
	case pqt.TypeSerial():
		return chooseType("int32", "*int32", "*int32", m)
	case pqt.TypeSerialSmall():
		return chooseType("int16", "*int16", "*int16", m)
	case pqt.TypeSerialBig():
		return chooseType("int64", "sql.NullInt64", "sql.NullInt64", m)
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		return chooseType("time.Time", "pq.NullTime", "pq.NullTime", m)
	case pqt.TypeReal():
		return chooseType("float32", "*float32", "*float32", m)
	case pqt.TypeDoublePrecision():
		return chooseType("float64", "sql.NullFloat64", "sql.NullFloat64", m)
	case pqt.TypeBytea(), pqt.TypeJSON(), pqt.TypeJSONB():
		return "[]byte"
	case pqt.TypeUUID():
		return chooseType("string", "sql.NullString", "sql.NullString", m)
	default:
		gt := t.String()
		switch {
		case strings.HasPrefix(gt, "SMALLINT["), strings.HasPrefix(gt, "INTEGER["), strings.HasPrefix(gt, "BIGINT["):
			return chooseType("pq.Int64Array", "NullInt64Array", "NullInt64Array", m)
		case strings.HasPrefix(gt, "DOUBLE PRECISION["):
			return chooseType("pq.Float64Array", "NullFloat64Array", "NullFloat64Array", m)
		case strings.HasPrefix(gt, "TEXT["):
			return chooseType("pq.StringArray", "NullStringArray", "NullStringArray", m)
		case strings.HasPrefix(gt, "DECIMAL"), strings.HasPrefix(gt, "NUMERIC"):
			return chooseType("float64", "sql.NullFloat64", "sql.NullFloat64", m)
		case strings.HasPrefix(gt, "VARCHAR"), strings.HasPrefix(gt, "CHARACTER"):
			return chooseType("string", "sql.NullString", "sql.NullString", m)
		default:
			return "interface{}"
		}
	}
}

func closeBrace(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprintln(w, "}")
	}
}
