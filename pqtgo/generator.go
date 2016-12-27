package pqtgo

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/piotrkowalczuk/pqt"
)

const (
	modeDefault = iota
	modeMandatory
	modeOptional
	modeCriteria

	dirtyAnd = `if com.Dirty {
		com.WriteString(" AND ")
	}
	com.Dirty = true
`
)

// Visibility ...
type Visibility string

const (
	// Public ...
	Public Visibility = "public"
	// Private ...
	Private Visibility = "private"
)

var keywords = map[string]string{
	"break":       "brk",
	"default":     "def",
	"func":        "fn",
	"interface":   "intf",
	"select":      "selec",
	"case":        "cas",
	"defer":       "defe",
	"go":          "g",
	"map":         "ma",
	"struct":      "struc",
	"chan":        "cha",
	"else":        "els",
	"goto":        "got",
	"package":     "pkg",
	"switch":      "switc",
	"const":       "cons",
	"fallthrough": "fallthroug",
	"if":          "i",
	"range":       "rang",
	"type":        "typ",
	"continue":    "cont",
	"for":         "fo",
	"import":      "impor",
	"return":      "rtn",
	"var":         "va",
}

// Generator ...
type Generator struct {
	ver      float32
	acronyms map[string]string
	imports  []string
	pkg      string
	vis      Visibility
}

// NewGenerator allocates new Generator.
func NewGenerator() *Generator {
	return &Generator{
		ver: 9.5,
		pkg: "main",
		vis: Private,
	}
}

// SetAcronyms ...
func (g *Generator) SetAcronyms(acronyms map[string]string) *Generator {
	g.acronyms = acronyms
	return g
}

// SetPostgresVersion sets version of postgres for which code will be generated.
func (g *Generator) SetPostgresVersion(ver float32) *Generator {
	g.ver = ver

	return g
}

// SetVisibility sets visibility that each struct, function, method or property will get.
func (g *Generator) SetVisibility(v Visibility) *Generator {
	g.vis = v

	return g
}

// SetImports ...
func (g *Generator) SetImports(imports ...string) *Generator {
	g.imports = imports

	return g
}

// AddImport ...
func (g *Generator) AddImport(i string) *Generator {
	if g.imports == nil {
		g.imports = make([]string, 0, 1)
	}

	g.imports = append(g.imports, i)
	return g
}

// SetPackage ...
func (g *Generator) SetPackage(pkg string) *Generator {
	g.pkg = pkg

	return g
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
func (g *Generator) GenerateTo(s *pqt.Schema, w io.Writer) error {
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
	for _, t := range s.Tables {
		g.generateConstants(b, t)
		g.generateColumns(b, t)
		g.generateEntity(b, t)
		g.generateEntityProp(b, t)
		g.generateEntityProps(b, t)
		g.generateIterator(b, t)
		g.generateCriteria(b, t)
		g.generateCriteriaWriteComposition(b, t)
		g.generatePatch(b, t)
		g.generateRepository(b, t)
	}

	return b, nil
}

func (g *Generator) generatePackage(code *bytes.Buffer) {
	fmt.Fprintf(code, "package %s\n", g.pkg)
}

func (g *Generator) generateImports(code *bytes.Buffer, schema *pqt.Schema) {
	imports := []string{
		"github.com/go-kit/kit/log",
		"github.com/m4rw3r/uuid",
	}
	imports = append(imports, g.imports...)
	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(CustomType); ok {
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
			}
		}
	}

	code.WriteString("import (\n")
	for _, imp := range imports {
		code.WriteRune('"')
		fmt.Fprint(code, imp)
		code.WriteRune('"')
		code.WriteRune('\n')
	}
	code.WriteString(")\n")
}

func (g *Generator) generateEntity(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "type %sEntity struct{\n", g.name(t.Name))
	for prop := range g.entityPropertiesGenerator(t) {
		fmt.Fprintf(w, "// %s ...\n", g.name(prop.Name))
		if prop.Tags != "" {
			fmt.Fprintf(w, "%s %s %s\n", g.name(prop.Name), prop.Type, prop.Tags)
		} else {
			fmt.Fprintf(w, "%s %s\n", g.name(prop.Name), prop.Type)
		}
	}
	fmt.Fprint(w, "}\n\n")
}

// entityPropertiesGenerator produces struct field definition for each column and relationship defined on a table.
// It thread differently relationship differently based on ownership.
func (g *Generator) entityPropertiesGenerator(t *pqt.Table) chan structField {
	fields := make(chan structField)

	go func(out chan structField) {
		for _, c := range t.Columns {
			if t := g.generateColumnTypeString(c, modeDefault); t != "<nil>" {
				out <- structField{Name: g.propertyName(c.Name), Type: t}
			}
		}

		for _, r := range t.OwnedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: g.propertyName(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.name(r.InversedTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: g.propertyName(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", g.name(r.InversedTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: g.propertyName(or(r.InversedName, r.InversedTable.Name)), Type: fmt.Sprintf("*%sEntity", g.name(r.InversedTable.Name))}
			}
		}

		for _, r := range t.InversedRelationships {
			switch r.Type {
			case pqt.RelationshipTypeOneToMany:
				out <- structField{Name: g.propertyName(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", g.name(r.OwnerTable.Name))}
			case pqt.RelationshipTypeOneToOne:
				out <- structField{Name: g.propertyName(or(r.OwnerName, r.OwnerTable.Name)), Type: fmt.Sprintf("*%sEntity", g.name(r.OwnerTable.Name))}
			case pqt.RelationshipTypeManyToOne:
				out <- structField{Name: g.propertyName(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.name(r.OwnerTable.Name))}
			}
		}

		for _, r := range t.ManyToManyRelationships {
			if r.Type != pqt.RelationshipTypeManyToMany {
				continue
			}

			switch {
			case r.OwnerTable == t:
				out <- structField{Name: g.propertyName(or(r.InversedName, r.InversedTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.name(r.InversedTable.Name))}
			case r.InversedTable == t:
				out <- structField{Name: g.propertyName(or(r.OwnerName, r.OwnerTable.Name+"s")), Type: fmt.Sprintf("[]*%sEntity", g.name(r.OwnerTable.Name))}
			}
		}

		close(out)
	}(fields)

	return fields
}

func (g *Generator) generateEntityProp(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "func (e *%sEntity) %s(cn string) (interface{}, bool) {\n", g.name(t.Name), g.name("Prop"))
	fmt.Fprintln(w, "switch cn {")
	for _, c := range t.Columns {
		fmt.Fprintf(w, "case %s:\n", g.columnNameWithTableName(t.Name, c.Name))
		if g.canBeNil(c, modeDefault) {
			fmt.Fprintf(w, "return e.%s, true\n", g.propertyName(c.Name))
		} else {
			fmt.Fprintf(w, "return &e.%s, true\n", g.propertyName(c.Name))
		}
	}
	fmt.Fprint(w, "default:\n")
	fmt.Fprint(w, "return nil, false\n")
	fmt.Fprint(w, "}\n}\n")
}

func (g *Generator) generateEntityProps(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "func (e *%sEntity) %s(cns ...string) ([]interface{}, error) {\n", g.name(t.Name), g.name("Props"))
	fmt.Fprintf(w, `
		res := make([]interface{}, 0, len(cns))
		for _, cn := range cns {
			if prop, ok := e.%s(cn); ok {
				res = append(res, prop)
			} else {
				return nil, fmt.Errorf("unexpected column provided: %%s", cn)
			}
		}
		return res, nil`, g.name("prop"))
	fmt.Fprint(w, "\n}\n")
}
func (g *Generator) generateIterator(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)
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
`, entityName, entityName, entityName, entityName, entityName, entityName, entityName, entityName, g.public(t.Name), entityName, g.public(t.Name), entityName, entityName, g.name("props"))
}

func (g *Generator) generateCriteria(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "type %sCriteria struct {\n", g.name(t.Name))
	fmt.Fprintf(w, "%s, %s int64\n", g.name("offset"), g.name("limit"))
	fmt.Fprintf(w, "%s map[string]bool\n", g.name("sort"))

ColumnLoop:
	for _, c := range t.Columns {
		if g.shouldBeColumnIgnoredForCriteria(c) {
			continue ColumnLoop
		}

		if t := g.generateColumnTypeString(c, modeCriteria); t != "<nil>" {
			fmt.Fprintf(w, "%s %s\n", g.propertyName(c.Name), t)
		}
	}
	fmt.Fprint(w, "}\n\n")
}

func (g *Generator) generatePatch(w io.Writer, t *pqt.Table) {
	fmt.Fprintf(w, "type %sPatch struct {\n", g.name(t.Name))

ArgumentsLoop:
	for _, c := range t.Columns {
		if c.PrimaryKey {
			continue ArgumentsLoop
		}

		if t := g.generateColumnTypeString(c, modeOptional); t != "<nil>" {
			fmt.Fprintf(w, "%s %s\n", g.propertyName(c.Name), t)
		}
	}
	fmt.Fprint(w, "}\n\n")
}

func (g *Generator) generateColumnTypeString(c *pqt.Column, m int32) string {
	switch m {
	case modeCriteria:
	case modeMandatory:
	case modeOptional:
	default:
		if c.NotNull || c.PrimaryKey {
			m = modeMandatory
		}
	}

	return g.generateType(c.Type, m)
}

func (g *Generator) generateType(t pqt.Type, m int32) string {
	switch tt := t.(type) {
	case pqt.MappableType:
		for _, mt := range tt.Mapping {
			return g.generateType(mt, m)
		}
		return ""
	case BuiltinType:
		return generateBuiltinType(tt, m)
	case pqt.BaseType:
		return generateBaseType(tt, m)
	case CustomType:
		return generateCustomType(tt, m)
	default:
		return ""
	}
}

func (g *Generator) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	g.generateConstantsColumns(code, table)
	g.generateConstantsConstraints(code, table)
	code.WriteString(")\n")
}

func (g *Generator) generateConstantsColumns(w io.Writer, table *pqt.Table) {
	fmt.Fprintf(w, `%s%s = "%s"
	`, g.name("table"), g.public(table.Name), table.FullName())

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(w, `%s%sColumn%s = "%s"
		`, g.name("table"), g.public(table.Name), g.public(name), name)
	}
}

func (g *Generator) generateConstantsConstraints(w io.Writer, table *pqt.Table) {
	for _, c := range tableConstraints(table) {
		name := fmt.Sprintf("%s", pqt.JoinColumns(c.Columns, "_"))
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			fmt.Fprintf(w, `%s%sConstraint%sCheck = "%s"`, g.name("table"), g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypePrimaryKey:
			fmt.Fprintf(w, `%s%sConstraintPrimaryKey = "%s"`, g.name("table"), g.public(table.Name), c.String())
		case pqt.ConstraintTypeForeignKey:
			fmt.Fprintf(w, `%s%sConstraint%sForeignKey = "%s"`, g.name("table"), g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeExclusion:
			fmt.Fprintf(w, `%s%sConstraint%sExclusion = "%s"`, g.name("table"), g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeUnique:
			fmt.Fprintf(w, `%s%sConstraint%sUnique = "%s"`, g.name("table"), g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeIndex:
			fmt.Fprintf(w, `%s%sConstraint%sIndex = "%s"`, g.name("table"), g.public(table.Name), g.public(name), c.String())
		}

		io.WriteString(w, "\n")
	}
}

func (g *Generator) generateColumns(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("var (\n")
	code.WriteString(g.name("table"))
	code.WriteString(g.public(table.Name))
	code.WriteString("Columns = []string{\n")

	for _, name := range sortedColumns(table.Columns) {
		g.writeTableNameColumnNameTo(code, table.Name, name)
		code.WriteRune(',')
		code.WriteRune('\n')
	}
	code.WriteString("}")
	code.WriteString(")\n")
}

func (g *Generator) generateRepository(b *bytes.Buffer, t *pqt.Table) {
	fmt.Fprintf(b, `
		type %sRepositoryBase struct {
			table string
			columns []string
			db *sql.DB
			dbg bool
			log log.Logger
		}
	`, g.name(t.Name))
	g.generateRepositoryScanRows(b, t)
	g.generateRepositoryCount(b, t)
	g.generateRepositoryFind(b, t)
	g.generateRepositoryFindIter(b, t)
	g.generateRepositoryFindOneByPrimaryKey(b, t)
	g.generateRepositoryFindOneByUniqueConstraint(b, t)
	g.generateRepositoryInsert(b, t)
	g.generateRepositoryUpsert(b, t)
	g.generateRepositoryUpdateOneByPrimaryKey(b, t)
	g.generateRepositoryUpdateOneByUniqueConstraint(b, t)
	g.generateRepositoryDeleteOneByPrimaryKey(b, t)
}

func (g *Generator) generateRepositoryFindPropertyQuery(w io.Writer, c *pqt.Column) {
	columnName := g.propertyName(c.Name)
	columnNameWithTable := g.columnNameWithTableName(c.Table.Name, c.Name)

	t := g.generateColumnTypeString(c, modeCriteria)
	if t == "<nil>" {
		return
	}
	if !g.generateRepositoryFindPropertyQueryByGoType(w, c, t, columnName, columnNameWithTable) {
		fmt.Fprintf(w, " if c.%s != nil {", g.propertyName(c.Name))
		fmt.Fprintf(w, dirtyAnd)
		fmt.Fprintf(w, `if _, err = com.WriteString(%s); err != nil {
			return
		}
		`, g.columnNameWithTableName(c.Table.Name, c.Name))
		fmt.Fprint(w, `if _, err = com.WriteString(" = "); err != nil {
			return
		}
		`)
		fmt.Fprintln(w, `if err = com.WritePlaceholder(); err != nil {
			return
		}
		`)
		fmt.Fprintln(w, `if com.Dirty {
			if opt.Cast != "" {
				if _, err = com.WriteString(opt.Cast); err != nil {
					return
				}
			} else {
				if _, err = com.WriteString(" "); err != nil {
					return
				}
			}
		}
		`)
		fmt.Fprintf(w, `com.Add(c.%s)
		}`, g.propertyName(c.Name))
	}
}

func (g *Generator) generateRepositoryFindPropertyQueryByGoType(w io.Writer, col *pqt.Column, goType, columnName, columnNameWithTable string) (done bool) {
	var isJSON bool
	switch col.Type {
	case pqt.TypeJSON(), pqt.TypeJSONB():
		isJSON = true
	}
	switch goType {
	case "uuid.UUID":
		fmt.Fprintf(w, `
			if !c.%s.IsZero() {
				%s
				if _, err = com.WriteString(%s); err != nil {
					return
				}
				if _, err = com.WriteString("!="); err != nil {
					return
				}
				if err = com.WritePlaceholder(); err != nil {
					return
				}
				com.Add(c.%s)
			}
		`, columnName, dirtyAnd, columnNameWithTable, columnName)
	case "*qtypes.Timestamp":
		fmt.Fprintf(w, `
				if c.%s != nil && c.%s.Valid {
					%st1 := c.%s.Value()
					if %st1 != nil {
						%s1, err := ptypes.Timestamp(%st1)
						if err != nil {
							return err
						}
						switch c.%s.Type {
						case qtypes.QueryType_NULL:
							%s
							com.WriteString(%s)
							if c.%s.Negation {
								com.WriteString(" IS NOT NULL ")
							} else {
								com.WriteString(" IS NULL ")
							}
						case qtypes.QueryType_EQUAL:
							%s
							com.WriteString(%s)
							if c.%s.Negation {
								com.WriteString(" <> ")
							} else {
								com.WriteString(" = ")
							}
							com.WritePlaceholder()
							com.Add(c.%s.Value())
						case qtypes.QueryType_GREATER:
							%s
							com.WriteString(%s)
							com.WriteString(">")
							com.WritePlaceholder()
							com.Add(c.%s.Value())
						case qtypes.QueryType_GREATER_EQUAL:
							%s
							com.WriteString(%s)
							com.WriteString(">=")
							com.WritePlaceholder()
							com.Add(c.%s.Value())
						case qtypes.QueryType_LESS:
							%s
							com.WriteString(%s)
							com.WriteString(" < ")
							com.WritePlaceholder()
							com.Add(c.%s.Value())
						case qtypes.QueryType_LESS_EQUAL:
							%s
							com.WriteString(%s)
							com.WriteString(" <= ")
							com.WritePlaceholder()
							com.Add(c.%s.Value())
						case qtypes.QueryType_IN:
							if len(c.%s.Values) >0 {
								%s
								com.WriteString(%s)
								com.WriteString(" IN (")
								for i, v := range c.%s.Values {
									if i != 0 {
										com.WriteString(", ")
									}
									com.WritePlaceholder()
									com.Add(v)
								}
								com.WriteString(") ")
							}
						case qtypes.QueryType_BETWEEN:
							%s
							%st2 := c.%s.Values[1]
							if %st2 != nil {
								%s2, err := ptypes.Timestamp(%st2)
								if err != nil {
									return err
								}
								com.WriteString(%s)
								com.WriteString(" > ")
								com.WritePlaceholder()
								com.Add(%s1)
								com.WriteString(" AND ")
								com.WriteString(%s)
								com.WriteString(" < ")
								com.WritePlaceholder()
								com.Add(%s2)
							}
						}
					}
				}
`,
			columnName, columnName,
			columnName, columnName,
			columnName,
			columnName,
			columnName, columnName,
			// NOT A NUMBER
			dirtyAnd,
			columnNameWithTable,
			columnName,
			// EQUAL
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// GREATER
			dirtyAnd,
			columnNameWithTable, columnName,
			// GREATER EQUAL
			dirtyAnd,
			columnNameWithTable, columnName,
			// LESS
			dirtyAnd,
			columnNameWithTable, columnName,
			// LESS EQUAL
			dirtyAnd,
			columnNameWithTable, columnName,
			// IN
			columnName,
			dirtyAnd,
			columnNameWithTable, columnName,
			// BETWEEN
			dirtyAnd,
			columnName, columnName,
			columnName, columnName,
			columnName,
			columnNameWithTable, columnName,
			columnNameWithTable, columnName,
		)
	case "*qtypes.Int64":
		fmt.Fprintf(w, `
		if err = pqtgo.WriteCompositionQueryInt64(c.%s, %s, com, &pqtgo.CompositionOpts{
		Joint: " AND ",
		IsJSON: %v,
	}); err != nil {
			return
		}`, columnName, columnNameWithTable, isJSON)
	case "*qtypes.Int32", "*qtypes.Float64":
		fmt.Fprintf(w, `
				if c.%s != nil && c.%s.Valid {
					switch c.%s.Type {
					case qtypes.QueryType_NULL:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" IS NOT NULL ")
						} else {
							com.WriteString(" IS NULL ")
						}
					case qtypes.QueryType_EQUAL:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" <> ")
						} else {
							com.WriteString(" = ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Value())
					case qtypes.QueryType_GREATER:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" <= ")
						} else {
							com.WriteString(" > ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Value())
					case qtypes.QueryType_GREATER_EQUAL:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" < ")
						} else {
							com.WriteString(" >= ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Value())
					case qtypes.QueryType_LESS:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" >= ")
						} else {
							com.WriteString(" < ")
						}
						com.WritePlaceholder()
						com.Add(c.%s)
					case qtypes.QueryType_LESS_EQUAL:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" > ")
						} else {
							com.WriteString(" <= ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Value())
					case qtypes.QueryType_IN:
						if len(c.%s.Values) >0 {
							%s
							com.WriteString(%s)
							if c.%s.Negation {
								com.WriteString(" NOT IN (")
							} else {
								com.WriteString(" IN (")
							}
							for i, v := range c.%s.Values {
								if i != 0 {
									com.WriteString(", ")
								}
								com.WritePlaceholder()
								com.Add(v)
							}
							com.WriteString(") ")
						}
					case qtypes.QueryType_BETWEEN:
						%s
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" <= ")
						} else {
							com.WriteString(" > ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Values[0])
						com.WriteString(" AND ")
						com.WriteString(%s)
						if c.%s.Negation {
							com.WriteString(" >= ")
						} else {
							com.WriteString(" < ")
						}
						com.WritePlaceholder()
						com.Add(c.%s.Values[1])
					}
				}
`,
			columnName, columnName,
			columnName,
			// NOT A NUMBER
			dirtyAnd,
			columnNameWithTable,
			columnName,
			// EQUAL
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// GREATER
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// GREATER EQUAL
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// LESS
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// LESS EQUAL
			dirtyAnd,
			columnNameWithTable, columnName, columnName,
			// IN
			columnName,
			dirtyAnd,
			columnNameWithTable,
			columnName, columnName,
			// BETWEEN
			dirtyAnd,
			columnNameWithTable,
			columnName, columnName,
			columnNameWithTable,
			columnName, columnName,
		)
	case "*qtypes.String":
		fmt.Fprintf(w, `
		if err = pqtgo.WriteCompositionQueryString(c.%s, %s, com, pqtgo.And); err != nil {
			return
		}`, columnName, columnNameWithTable)
	default:
		if strings.HasPrefix(goType, "*ntypes.") {
			fmt.Fprintf(w, " if c.%s != nil && c.%s.Valid {", columnName, columnName)
			fmt.Fprintf(w, dirtyAnd)
			fmt.Fprintf(w, "com.WriteString(%s)\n", columnNameWithTable)
			fmt.Fprintln(w, `com.WriteString(" = ")`)
			fmt.Fprintln(w, `com.WritePlaceholder()`)
			fmt.Fprintf(w, `com.Add(c.%s)
		}`, columnName)
			return true
		}
		return
	}
	return true
}

func (g *Generator) generateRepositoryFindSingleExpression(w io.Writer, c *pqt.Column) {
	if mappt, ok := c.Type.(pqt.MappableType); ok {
	MappingLoop:
		for _, mt := range mappt.Mapping {
			switch mtt := mt.(type) {
			case CustomType:
				if gct := generateCustomType(mtt, modeCriteria); strings.HasPrefix(gct, "*qtypes.") {
					columnName := g.propertyName(c.Name)
					columnNameWithTable := g.columnNameWithTableName(c.Table.Name, c.Name)

					if !g.generateRepositoryFindPropertyQueryByGoType(w, c, gct, columnName, columnNameWithTable) {
						panic("custom type criteria variant cannot be generated")
					}
					break MappingLoop
				}

				if mtt.criteriaTypeOf == nil {
					fmt.Printf("%s.%s: criteria type of nil\n", c.Table.FullName(), c.Name)
					return
				}
				if mtt.criteriaTypeOf.Kind() == reflect.Invalid {
					fmt.Printf("%s.%s: criteria invalid type\n", c.Table.FullName(), c.Name)
					return
				}

				columnName := g.propertyName(c.Name)
				columnNameWithTable := g.columnNameWithTableName(c.Table.Name, c.Name)
				zero := reflect.Zero(mtt.criteriaTypeOf)
				// Checks if custom type implements Criterion interface.
				// If it's true then just use it.
				if zero.CanInterface() {
					if _, ok := zero.Interface().(CompositionWriter); ok {
						if zero.IsNil() {
							fmt.Fprintf(w, "if c.%s != nil {", columnName)
						}
						fmt.Fprintf(w, `
							if err = c.%s.WriteComposition(%s, com, pqtgo.And); err != nil {
								return
							}
						`, columnName, columnNameWithTable)
						if zero.IsNil() {
							fmt.Fprintln(w, "}")
						}
						return
					}
				}

				switch zero.Kind() {
				case reflect.Map:
					// TODO: implement
					return
				case reflect.Struct:
					for i := 0; i < zero.NumField(); i++ {
						field := zero.Field(i)
						fieldName := columnName + "." + zero.Type().Field(i).Name
						fieldJSONName := strings.Split(zero.Type().Field(i).Tag.Get("json"), ", ")[0]
						columnNameWithTableAndJSONSelector := fmt.Sprintf(`%s + " -> '%s'"`, columnNameWithTable, fieldJSONName)

						// If struct is nil, it's properties should not be accessed.
						fmt.Fprintf(w, `if c.%s != nil {
						`, g.propertyName(c.Name))
						g.generateRepositoryFindPropertyQueryByGoType(w, c, field.Type().String(), fieldName, columnNameWithTableAndJSONSelector)
						fmt.Fprintln(w, "}")
					}
				}
			default:
				g.generateRepositoryFindPropertyQuery(w, c)
			}
		}
	} else {
		g.generateRepositoryFindPropertyQuery(w, c)
	}
	fmt.Fprintln(w, "")
}

func (g *Generator) generateCriteriaWriteComposition(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)
	fmt.Fprintf(w, `func (c *%sCriteria) WriteComposition(sel string, com *pqtgo.Composer, opt *pqtgo.CompositionOpts) (err error) {
	`, entityName) // It's probably not enough but its good start.
	for _, c := range t.Columns {
		if g.shouldBeColumnIgnoredForCriteria(c) {
			continue
		}

		g.generateRepositoryFindSingleExpression(w, c)
	}
	fmt.Fprintf(w, `
	if len(c.%s) > 0 {
		i:=0
		com.WriteString(" ORDER BY ")

		for cn, asc := range c.%s {
			for _, tcn := range %s%sColumns {
				if cn == tcn {
					if i > 0 {
						com.WriteString(", ")
					}
					com.WriteString(cn)
					if !asc {
						com.WriteString(" DESC ")
					}
					i++
					break
				}
			}
		}
	}
	if c.%s > 0 {
		if _, err = com.WriteString(" OFFSET "); err != nil {
			return
		}
		if err = com.WritePlaceholder(); err != nil {
			return
		}
		if _, err = com.WriteString(" "); err != nil {
			return
		}
		com.Add(c.%s)
	}
	if c.%s > 0 {
		if _, err = com.WriteString(" LIMIT "); err != nil {
			return
		}
		if err = com.WritePlaceholder(); err != nil {
			return
		}
		if _, err = com.WriteString(" "); err != nil {
			return
		}
		com.Add(c.%s)
	}

	return
}
`, g.name("sort"), g.name("sort"),
		g.name("table"), g.public(t.Name),
		g.name("offset"), g.name("offset"),
		g.name("limit"), g.name("limit"))
}

func (g *Generator) generateRepositoryScanRows(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)
	fmt.Fprintf(w, `func %s%sRows(rows *sql.Rows) ([]*%sEntity, error) {
	`, g.name("Scan"), g.public(t.Name), entityName)
	fmt.Fprintf(w, `var (
		entities []*%sEntity
		err error
	)
	for rows.Next() {
		var ent %sEntity
		err = rows.Scan(
	`, entityName, entityName)
	for _, c := range t.Columns {
		fmt.Fprintf(w, "&ent.%s,\n", g.propertyName(c.Name))
	}
	fmt.Fprint(w, `)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &ent)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		return entities, nil
	}

	`)
}

func (g *Generator) generateRepositoryFindBody(w io.Writer, t *pqt.Table) {
	fmt.Fprint(w, `
	com := pqtgo.NewComposer(1)
	buf := bytes.NewBufferString("SELECT ")
	buf.WriteString(strings.Join(r.columns, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(r.table)
	buf.WriteString(" ")

	if err := c.WriteComposition("", com, pqtgo.And); err != nil {
		return nil, err
	}
	if com.Dirty {
		buf.WriteString(" WHERE ")
	}
	if com.Len() > 0 {
		buf.ReadFrom(com)
	}

	if r.dbg {
		if err := r.log.Log("msg", buf.String(), "function", "Find"); err != nil {
			return nil, err
		}
	}

	rows, err := r.db.Query(buf.String(), com.Args()...)
	if err != nil {
		return nil, err
	}
`)
}

func (g *Generator) generateRepositoryFind(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)

	fmt.Fprintf(w, `
func (r *%sRepositoryBase) %s(c *%sCriteria) ([]*%sEntity, error) {
`, entityName, g.name("Find"), entityName, entityName)
	g.generateRepositoryFindBody(w, t)
	fmt.Fprintf(w, `
	defer rows.Close()

	return %s%sRows(rows)
}
`, g.name("Scan"), g.public(t.Name))
}

func (g *Generator) generateRepositoryFindIter(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(c *%sCriteria) (*%sIterator, error) {
`, entityName, g.name("FindIter"), entityName, entityName)
	g.generateRepositoryFindBody(w, t)
	fmt.Fprintf(w, `

	return &%sIterator{rows: rows}, nil
}
`, g.name(t.Name))
}

func (g *Generator) generateRepositoryCount(w io.Writer, t *pqt.Table) {
	entityName := g.name(t.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(c *%sCriteria) (int64, error) {
`, entityName, g.name("count"), entityName)
	fmt.Fprintf(w, `
	com := pqtgo.NewComposer(%d)
	buf := bytes.NewBufferString("SELECT COUNT(*) FROM ")
	buf.WriteString(r.table)

	if err := c.WriteComposition("", com, pqtgo.And); err != nil {
		return 0, err
	}
	if com.Dirty {
		buf.WriteString(" WHERE ")
	}
	if com.Len() > 0 {
		buf.ReadFrom(com)
	}

	if r.dbg {
		if err := r.log.Log("msg", buf.String(), "function", "Count"); err != nil {
			return 0, err
		}
	}

	var count int64
	if err := r.db.QueryRow(buf.String(), com.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
`, len(t.Columns))
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.name(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(code, `func (r *%sRepositoryBase) %s%s(%s %s) (*%sEntity, error) {`,
		entityName,
		g.name("FindOneBy"),
		g.public(pk.Name),
		g.private(pk.Name),
		g.generateColumnTypeString(pk, modeMandatory),
		entityName,
	)
	fmt.Fprintf(code, `var (
		ent %sEntity
	)`, entityName)
	code.WriteRune('\n')
	fmt.Fprint(code, "query := `SELECT ")
	for i, c := range table.Columns {
		fmt.Fprintf(code, "%s", c.Name)
		if i != len(table.Columns)-1 {
			code.WriteRune(',')
		}
		code.WriteRune('\n')
	}
	fmt.Fprintf(code, " FROM %s WHERE %s = $1`", table.FullName(), pk.Name)

	fmt.Fprintf(code, `
	err := r.db.QueryRow(query, %s).Scan(
	`, g.private(pk.Name))
	for _, c := range table.Columns {
		fmt.Fprintf(code, "&ent.%s,\n", g.propertyName(c.Name))
	}
	fmt.Fprint(code, `)
		if err != nil {
			return nil, err
		}

		return &ent, nil
}
`)
}

func (g *Generator) generateRepositoryFindOneByUniqueConstraint(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.name(table.Name)
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
		arguments := ""
		methodName := "FindOneBy"
		for i, c := range u.Columns {
			if i != 0 {
				methodName += "And"
				arguments += ", "
			}
			methodName += g.public(c.Name)
			arguments += fmt.Sprintf("%s %s", g.private(columnForeignName(c)), g.generateColumnTypeString(c, modeMandatory))
		}
		fmt.Fprintf(code, `func (r *%sRepositoryBase) %s(%s) (*%sEntity, error) {`, entityName, g.name(methodName), arguments, entityName)
		fmt.Fprintf(code, `var (
			ent %sEntity
		)`, entityName)
		code.WriteRune('\n')
		fmt.Fprint(code, "query := `SELECT ")
		for i, c := range table.Columns {
			if i != 0 {
				code.WriteString(", ")
			}
			fmt.Fprintf(code, "%s", c.Name)
		}
		fmt.Fprintf(code, " FROM %s WHERE ", table.FullName())
		for i, c := range u.Columns {
			if i != 0 {
				fmt.Fprint(code, " AND ")
			}
			fmt.Fprintf(code, "%s = $%d", c.Name, i+1)
		}
		fmt.Fprintln(code, "`")

		fmt.Fprint(code, "err := r.db.QueryRow(query, ")
		for i, c := range u.Columns {
			if i != 0 {
				fmt.Fprint(code, ", ")
			}
			fmt.Fprintf(code, "%s", g.private(columnForeignName(c)))
		}
		fmt.Fprint(code, ").Scan(\n")
		for _, c := range table.Columns {
			fmt.Fprintf(code, "&ent.%s,\n", g.propertyName(c.Name))
		}
		fmt.Fprint(code, `)
			if err != nil {
				return nil, err
			}

			return &ent, nil
	}
	`)

	}
}

func (g *Generator) generateRepositoryInsert(w io.Writer,
	table *pqt.Table) {
	entityName := g.name(table.Name)

	fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(e *%sEntity) (*%sEntity, error) {`, entityName, g.name("Insert"), entityName, entityName)
	fmt.Fprintf(w, `
		insert := pqcomp.New(0, %d)
	`, len(table.Columns))

ColumnsLoop:
	for _, c := range table.Columns {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue ColumnsLoop
		default:
			if g.canBeNil(c, modeMandatory) {
				fmt.Fprintf(w, `
					if e.%s != nil {
						insert.AddExpr(%s, "", e.%s)
					}
				`,
					g.propertyName(c.Name),
					g.columnNameWithTableName(table.Name, c.Name),
					g.propertyName(c.Name),
				)
			} else {
				fmt.Fprintf(
					w,
					`insert.AddExpr(%s, "", e.%s)`,
					g.columnNameWithTableName(table.Name, c.Name),
					g.propertyName(c.Name),
				)
			}
			fmt.Fprintln(w, "")
		}
	}
	fmt.Fprint(w, `
		b := bytes.NewBufferString("INSERT INTO " + r.table)

		if insert.Len() != 0 {
			b.WriteString(" (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.Key())
			}
			insert.Reset()
			b.WriteString(") VALUES (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.PlaceHolder())
			}
			b.WriteString(")")
			if len(r.columns) > 0 {
				b.WriteString(" RETURNING ")
				b.WriteString(strings.Join(r.columns, ", "))
			}
		}

		if r.dbg {
			if err := r.log.Log("msg", b.String(), "function", "Insert"); err != nil {
				return nil, err
			}
		}

		err := r.db.QueryRow(b.String(), insert.Args()...).Scan(
	`)

	for _, c := range table.Columns {
		fmt.Fprintf(w, "&e.%s,\n", g.propertyName(c.Name))
	}
	fmt.Fprint(w, `)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
`)
}

func (g *Generator) generateRepositoryUpsert(code *bytes.Buffer, table *pqt.Table) {
	if g.ver < 9.5 {
		return
	}
	entityName := g.name(table.Name)

	fmt.Fprintf(code, `func (r *%sRepositoryBase) %s(e *%sEntity, p *%sPatch, inf ...string) (*%sEntity, error) {`,
		entityName, g.name("Upsert"),
		entityName, entityName, entityName,
	)
	fmt.Fprintf(code, `
		insert := pqcomp.New(0, %d)
		update := insert.Compose(%d)
	`, len(table.Columns), len(table.Columns))

InsertLoop:
	for _, c := range table.Columns {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue InsertLoop
		default:
			if g.canBeNil(c, modeMandatory) {
				fmt.Fprintf(code, `
					if e.%s != nil {
						insert.AddExpr(%s, "", e.%s)
					}
				`,
					g.propertyName(c.Name),
					g.columnNameWithTableName(table.Name, c.Name), g.propertyName(c.Name),
				)
			} else {
				fmt.Fprintf(code, `insert.AddExpr(%s, "", e.%s)`,
					g.columnNameWithTableName(table.Name, c.Name),
					g.propertyName(c.Name),
				)
			}
			fmt.Fprintln(code, "")
		}
	}
	fmt.Fprintln(code, "if len(inf) > 0 {")
UpdateLoop:
	for _, c := range table.Columns {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue UpdateLoop
		default:
			if g.canBeNil(c, modeOptional) {
				fmt.Fprintf(code, `
					if p.%s != nil {
						update.AddExpr(%s, "=", p.%s)
					}
				`,
					g.propertyName(c.Name),
					g.columnNameWithTableName(table.Name, c.Name),
					g.propertyName(c.Name),
				)
			} else {
				fmt.Fprintf(code, `update.AddExpr(%s, "=", p.%s)`, g.columnNameWithTableName(table.Name, c.Name), g.propertyName(c.Name))
			}
			fmt.Fprintln(code, "")
		}
	}
	fmt.Fprintln(code, "}")

	fmt.Fprint(code, `
		b := bytes.NewBufferString("INSERT INTO " + r.table)

		if insert.Len() > 0 {
			b.WriteString(" (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.Key())
			}
			insert.Reset()
			b.WriteString(") VALUES (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.PlaceHolder())
			}
			b.WriteString(")")
		}
		b.WriteString(" ON CONFLICT ")
		if len(inf) > 0 && update.Len() > 0 {
			b.WriteString(" (")
			for j, i := range inf {
				if j != 0 {
					b.WriteString(", ")
				}
				b.WriteString(i)
			}
			b.WriteString(") ")
			b.WriteString(" DO UPDATE SET ")
			for update.Next() {
				if !update.First() {
					b.WriteString(", ")
				}

				b.WriteString(update.Key())
				b.WriteString(" ")
				b.WriteString(update.Oper())
				b.WriteString(" ")
				b.WriteString(update.PlaceHolder())
			}
		} else {
			b.WriteString(" DO NOTHING ")
		}
		if insert.Len() > 0 {
			if len(r.columns) > 0 {
				b.WriteString(" RETURNING ")
				b.WriteString(strings.Join(r.columns, ", "))
			}
		}

		if r.dbg {
			if err := r.log.Log("msg", b.String(), "function", "Upsert"); err != nil {
				return nil, err
			}
		}

		err := r.db.QueryRow(b.String(), insert.Args()...).Scan(
	`)

	for _, c := range table.Columns {
		fmt.Fprintf(code, "&e.%s,\n", g.propertyName(c.Name))
	}
	fmt.Fprint(code, `)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
`)
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraint(w io.Writer, table *pqt.Table) {
	entityName := g.name(table.Name)
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
		arguments := ""
		methodName := "UpdateOneBy"
		for i, c := range u.Columns {
			if i != 0 {
				methodName += "And"
				arguments += ", "
			}
			methodName += g.public(c.Name)
			arguments += fmt.Sprintf("%s %s", g.private(columnForeignName(c)), g.generateColumnTypeString(c, modeMandatory))
		}
		fmt.Fprintf(w, `func (r *%sRepositoryBase) %s(%s, patch *%sPatch) (*%sEntity, error) {
		`, entityName, g.name(methodName), arguments, entityName, entityName)
		fmt.Fprintf(w, "update := pqcomp.New(%d, %d)\n", len(u.Columns), len(table.Columns))
		for _, c := range u.Columns {
			fmt.Fprintf(w, "update.AddArg(%s)\n", g.private(columnForeignName(c)))
		}
		pk, pkOK := table.PrimaryKey()
	ColumnsLoop:
		for _, c := range table.Columns {
			if pkOK && c == pk {
				continue ColumnsLoop
			}
			for _, uc := range u.Columns {
				if c == uc {
					continue
				}
			}
			if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
				switch c.Type {
				case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
					fmt.Fprintf(w, "if patch.%s != nil {\n", g.propertyName(c.Name))

				}
			} else if g.canBeNil(c, modeOptional) {
				fmt.Fprintf(w, "if patch.%s != nil {\n", g.propertyName(c.Name))
			}

			fmt.Fprint(w, "update.AddExpr(")
			g.writeTableNameColumnNameTo(w, c.Table.Name, c.Name)
			fmt.Fprintf(w, ", pqcomp.Equal, patch.%s)\n", g.propertyName(c.Name))

			if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
				switch c.Type {
				case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
					fmt.Fprint(w, `} else {`)
					fmt.Fprint(w, "update.AddExpr(")
					g.writeTableNameColumnNameTo(w, c.Table.Name, c.Name)
					fmt.Fprintf(w, `, pqcomp.Equal, "%s")`, d)
				}
			}
			if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
				switch c.Type {
				case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
					fmt.Fprint(w, "\n}\n")
				}
			} else if g.canBeNil(c, modeOptional) {
				fmt.Fprint(w, "\n}\n")
			}
		}

		fmt.Fprintf(w, `
	if update.Len() == 0 {
		return nil, errors.New("%s update failure, nothing to update")
	}`, entityName)

		fmt.Fprintf(w, `
	query := "UPDATE %s SET "
	for update.Next() {
		if !update.First() {
			query += ", "
		}

		query += update.Key() + " " + update.Oper() + " " + update.PlaceHolder()
	}
`, table.FullName())
		fmt.Fprint(w, `query += " WHERE `)
		for i, c := range u.Columns {
			if i != 0 {
				fmt.Fprint(w, " AND ")
			}
			fmt.Fprintf(w, "%s = $%d", c.Name, i+1)
		}
		fmt.Fprintf(w, ` RETURNING " + strings.Join(r.columns, ", ")
	if r.dbg {
		if err := r.log.Log("msg", query, "function", "%s"); err != nil {
			return nil, err
		}
	}
	var e %sEntity
	err := r.db.QueryRow(query, update.Args()...).Scan(
	`, methodName, entityName)
		for _, c := range table.Columns {
			fmt.Fprintf(w, "&e.%s,\n", g.propertyName(c.Name))
		}
		fmt.Fprint(w, `)
if err != nil {
	return nil, err
}


return &e, nil
}
`)
	}
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKey(w io.Writer, table *pqt.Table) {
	entityName := g.name(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(w, "func (r *%sRepositoryBase) %s%s(%s %s, patch *%sPatch) (*%sEntity, error) {\n", entityName, g.name("UpdateOneBy"), g.public(pk.Name), g.private(pk.Name), g.generateColumnTypeString(pk, modeMandatory), entityName, entityName)
	fmt.Fprintf(w, "update := pqcomp.New(1, %d)\n", len(table.Columns))
	fmt.Fprintf(w, "update.AddArg(%s)\n", g.private(pk.Name))
	fmt.Fprintln(w, "")

ColumnsLoop:
	for _, c := range table.Columns {
		if c == pk {
			continue ColumnsLoop
		}
		if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprintf(w, "if patch.%s != nil {\n", g.propertyName(c.Name))

			}
		} else if g.canBeNil(c, modeOptional) {
			fmt.Fprintf(w, "if patch.%s != nil {\n", g.propertyName(c.Name))
		}

		fmt.Fprint(w, "update.AddExpr(")
		g.writeTableNameColumnNameTo(w, c.Table.Name, c.Name)
		fmt.Fprintf(w, ", pqcomp.Equal, patch.%s)\n", g.propertyName(c.Name))

		if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprint(w, `} else {`)
				fmt.Fprint(w, "update.AddExpr(")
				g.writeTableNameColumnNameTo(w, c.Table.Name, c.Name)
				fmt.Fprintf(w, `, pqcomp.Equal, "%s")`, d)
			}
		}
		if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprint(w, "\n}\n")
			}
		} else if g.canBeNil(c, modeOptional) {
			fmt.Fprint(w, "\n}\n")
		}
	}
	fmt.Fprintf(w, `
	if update.Len() == 0 {
		return nil, errors.New("%s update failure, nothing to update")
	}`, entityName)

	fmt.Fprintf(w, `
	query := "UPDATE %s SET "
	for update.Next() {
		if !update.First() {
			query += ", "
		}

		query += update.Key() + " " + update.Oper() + " " + update.PlaceHolder()
	}
	query += " WHERE %s = $1 RETURNING " + strings.Join(r.columns, ", ")
	var e %sEntity
	err := r.db.QueryRow(query, update.Args()...).Scan(
	`, table.FullName(), pk.Name, entityName)
	for _, c := range table.Columns {
		fmt.Fprintf(w, "&e.%s,\n", g.propertyName(c.Name))
	}
	fmt.Fprint(w, `)
if err != nil {
	return nil, err
}


return &e, nil
}
`)
}

func (g *Generator) generateRepositoryDeleteOneByPrimaryKey(code *bytes.Buffer,
	table *pqt.Table) {
	entityName := g.name(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(code, `
		func (r *%sRepositoryBase) %s%s(%s %s) (int64, error) {
			query := "DELETE FROM %s WHERE %s = $1"

			res, err := r.db.Exec(query, %s)
			if err != nil {
				return 0, err
			}

			return res.RowsAffected()
		}
`, entityName, g.name("DeleteOneBy"), g.public(pk.Name), g.private(pk.Name), g.generateColumnTypeString(pk, 1), table.FullName(), pk.Name, g.private(pk.Name))
}

func sortedColumns(columns []*pqt.Column) []string {
	tmp := make([]string, 0, len(columns))
	for _, c := range columns {
		tmp = append(tmp, c.Name)
	}
	sort.Strings(tmp)

	return tmp
}

func snake(s string, private bool, acronyms map[string]string) string {
	var parts []string
	parts1 := strings.Split(s, "_")
	for _, p1 := range parts1 {
		parts2 := strings.Split(p1, "/")
		for _, p2 := range parts2 {
			parts3 := strings.Split(p2, "-")
			parts = append(parts, parts3...)
		}
	}

	for i, part := range parts {
		if !private || i > 0 {
			if formatted, ok := acronyms[part]; ok {
				parts[i] = formatted

				continue
			}
		}

		parts[i] = xstrings.FirstRuneToUpper(part)
	}

	if private {
		parts[0] = xstrings.FirstRuneToLower(parts[0])
	}
	return strings.Join(parts, "")
}

func (g *Generator) propertyName(s string) string {
	n := g.name(s)
	if r, ok := keywords[n]; ok {
		return r
	}

	return n
}

func (g *Generator) name(s string) string {
	return snake(s, g.vis == Private, g.acronyms)
}

func (g *Generator) private(s string) string {
	return snake(s, true, g.acronyms)
}

func (g *Generator) public(s string) string {
	return snake(s, false, g.acronyms)
}

func (g *Generator) isStruct(c *pqt.Column, m int) bool {
	if tp, ok := c.Type.(pqt.MappableType); ok {
		for _, mapto := range tp.Mapping {
			if ct, ok := mapto.(CustomType); ok {
				switch m {
				case modeMandatory:
					return ct.mandatoryTypeOf.Kind() == reflect.Struct
				case modeOptional:
					return ct.optionalTypeOf.Kind() == reflect.Struct
				case modeCriteria:
					return ct.criteriaTypeOf.Kind() == reflect.Struct
				default:
					return false
				}
			}
		}
	}
	return false
}

func (g *Generator) canBeNil(c *pqt.Column, m int) bool {
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
	return false
}

func chooseType(tm, to, tc string, mode int32) string {
	switch mode {
	case modeCriteria:
		return tc
	case modeMandatory:
		return tm
	case modeOptional:
		return to
	case modeDefault:
		return to
	default:
		panic("unknown mode")
	}
}

func tableConstraints(t *pqt.Table) []*pqt.Constraint {
	var constraints []*pqt.Constraint
	for _, c := range t.Columns {
		constraints = append(constraints, c.Constraints()...)
	}

	return append(constraints, t.Constraints...)
}

func generateBaseType(t pqt.Type, m int32) string {
	switch t {
	case pqt.TypeText():
		return chooseType("string", "*ntypes.String", "*qtypes.String", m)
	case pqt.TypeBool():
		return chooseType("bool", "*ntypes.Bool", "*ntypes.Bool", m)
	case pqt.TypeIntegerSmall():
		return chooseType("int16", "*int16", "*int16", m)
	case pqt.TypeInteger():
		return chooseType("int32", "*ntypes.Int32", "*ntypes.Int32", m)
	case pqt.TypeIntegerBig():
		return chooseType("int64", "*ntypes.Int64", "*qtypes.Int64", m)
	case pqt.TypeSerial():
		return chooseType("int32", "*ntypes.Int32", "*ntypes.Int32", m)
	case pqt.TypeSerialSmall():
		return chooseType("int16", "*int16", "*int16", m)
	case pqt.TypeSerialBig():
		return chooseType("int64", "*ntypes.Int64", "*qtypes.Int64", m)
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		return chooseType("time.Time", "*time.Time", "*qtypes.Timestamp", m)
	case pqt.TypeReal():
		return chooseType("float32", "*ntypes.Float32", "*ntypes.Float32", m)
	case pqt.TypeDoublePrecision():
		return chooseType("float64", "*ntypes.Float64", "*qtypes.Float64", m)
	case pqt.TypeBytea(), pqt.TypeJSON(), pqt.TypeJSONB():
		return "[]byte"
	case pqt.TypeUUID():
		return "uuid.UUID"
	default:
		gt := t.String()
		switch {
		case strings.HasPrefix(gt, "SMALLINT["):
			return chooseType("pq.Int64Array", "pq.Int64Array", "*qtypes.Int64", m)
		case strings.HasPrefix(gt, "INTEGER["):
			return chooseType("pq.Int64Array", "pq.Int64Array", "*qtypes.Int64", m)
		case strings.HasPrefix(gt, "BIGINT["):
			return chooseType("pq.Int64Array", "pq.Int64Array", "*qtypes.Int64", m)
		case strings.HasPrefix(gt, "DOUBLE PRECISION["):
			return chooseType("pq.Float64Array", "pq.Float64Array", "*qtypes.Float64", m)
		case strings.HasPrefix(gt, "TEXT["):
			return "pq.StringArray"
		case strings.HasPrefix(gt, "DECIMAL"), strings.HasPrefix(gt, "NUMERIC"):
			return chooseType("float64", "*ntypes.Float64", "*qtypes.Float64", m)
		case strings.HasPrefix(gt, "VARCHAR"):
			return chooseType("string", "*ntypes.String", "*qtypes.String", m)
		default:
			return "interface{}"
		}
	}
}

func generateBuiltinType(t BuiltinType, m int32) (r string) {
	switch types.BasicKind(t) {
	case types.Bool:
		r = chooseType("bool", "*ntypes.Bool", "*ntypes.Bool", m)
	case types.Int:
		r = chooseType("int", "*ntypes.Int", "*qtypes.Int64", m)
	case types.Int8:
		r = chooseType("int8", "*int8", "*int8", m)
	case types.Int16:
		r = chooseType("int16", "*int16", "*int16", m)
	case types.Int32:
		r = chooseType("int32", "*ntypes.Int32", "*qtypes.Int64", m)
	case types.Int64:
		r = chooseType("int64", "*ntypes.Int64", "*qtypes.Int64", m)
	case types.Uint:
		r = chooseType("uint", "*uint", "*uint", m)
	case types.Uint8:
		r = chooseType("uint8", "*uint8", "*uint8", m)
	case types.Uint16:
		r = chooseType("uint16", "*uint16", "*uint16", m)
	case types.Uint32:
		r = chooseType("uint32", "*ntypes.Uint32", "*ntypes.Uint32", m)
	case types.Uint64:
		r = chooseType("uint64", "*uint64", "*uint64", m)
	case types.Float32:
		r = chooseType("float32", "*ntypes.Float32", "*qtypes.Float64", m)
	case types.Float64:
		r = chooseType("float64", "*ntypes.Float64", "*qtypes.Float64", m)
	case types.Complex64:
		r = chooseType("complex64", "*complex64", "*complex64", m)
	case types.Complex128:
		r = chooseType("complex128", "*complex128", "*complex128", m)
	case types.String:
		r = chooseType("string", "*ntypes.String", "*qtypes.String", m)
	default:
		r = "invalid"
	}

	return
}

func generateCustomType(t CustomType, m int32) string {
	goType := func(tp reflect.Type) string {
		if tp == nil {
			return "<nil>"
		}
		if tp.Kind() == reflect.Struct {
			return "*" + tp.String()
		}
		return tp.String()
	}
	return chooseType(
		goType(t.mandatoryTypeOf),
		goType(t.optionalTypeOf),
		goType(t.criteriaTypeOf),
		m,
	)
}

func (g *Generator) writeTableNameColumnNameTo(w io.Writer, tableName, columnName string) {
	fmt.Fprintf(w, "%s%sColumn%s", g.name("table"), g.public(tableName), g.public(columnName))
}

func (g *Generator) columnNameWithTableName(tableName, columnName string) string {
	return fmt.Sprintf("%s%sColumn%s", g.name("table"), g.public(tableName), g.public(columnName))
}

func (g *Generator) shouldBeColumnIgnoredForCriteria(c *pqt.Column) bool {
	return false
	//if mt, ok := c.Type.(pqt.MappableType); ok {
	//	switch mt.From {
	//	case pqt.TypeJSON(), pqt.TypeJSONB():
	//		for _, to := range mt.Mapping {
	//			if ct, ok := to.(*CustomType); ok {
	//				switch ct.valueOf.Kind() {
	//				case reflect.Struct:
	//					return false
	//				case reflect.Map:
	//					return false
	//				case reflect.Slice:
	//					return false
	//				case reflect.Slice:
	//					return false
	//				}
	//			}
	//		}
	//		return true
	//	}
	//}
	//
	//return false
}

type structField struct {
	Name string
	Type string
	Tags reflect.StructTag
}

func or(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
}

func columnForeignName(c *pqt.Column) string {
	return c.Table.Name + "_" + c.Name
}
