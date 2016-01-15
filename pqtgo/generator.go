package pqtgo

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"sort"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/piotrkowalczuk/pqt"
)

type generator struct {
	acronyms map[string]string
	imports  []string
	pkg      string
}

// Generator ...
func Generator() *generator {
	return &generator{
		pkg: "main",
	}
}

// SetAcronyms ...
func (g *generator) SetAcronyms(acronyms map[string]string) *generator {
	g.acronyms = acronyms

	return g
}

// SetImports ...
func (g *generator) SetImports(imports ...string) *generator {
	g.imports = imports

	return g
}

// AddImport ...
func (g *generator) AddImport(i string) *generator {
	if g.imports == nil {
		g.imports = make([]string, 0, 1)
	}

	g.imports = append(g.imports, i)
	return g
}

// SetPackage ...
func (g *generator) SetPackage(pkg string) *generator {
	g.pkg = pkg

	return g
}

// Generate ...
func (g *generator) Generate(s *pqt.Schema) ([]byte, error) {
	code, err := g.generate(s)
	if err != nil {
		return nil, err
	}

	return code.Bytes(), nil
}

// GenerateTo ...
func (g *generator) GenerateTo(s *pqt.Schema, w io.Writer) error {
	code, err := g.generate(s)
	if err != nil {
		return err
	}

	_, err = code.WriteTo(w)
	return err
}

func (g *generator) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	code := bytes.NewBuffer(nil)

	g.generatePackage(code)
	g.generateImports(code, s)
	for _, table := range s.Tables {
		g.generateConstants(code, table)
		g.generateColumns(code, table)
		g.generateEntity(code, table)
	}

	return code, nil
}

func (g *generator) generatePackage(code *bytes.Buffer) {
	fmt.Fprintf(code, "package %s \n", g.pkg)
}

func (g *generator) generateImports(code *bytes.Buffer, schema *pqt.Schema) {
	imports := []string{}

	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(CustomType); ok {
				imports = append(imports, ct.pkg)
			}
		}
	}

	code.WriteString("import (\n")
	for _, imp := range imports {
		fmt.Fprintf(code, `"%s"`, imp)
	}
	code.WriteString(")\n")
}

func (g *generator) generateEntity(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("type " + g.private(table.Name) + "Entity struct {")
	for _, c := range table.Columns {
		code.WriteString(g.public(c.Name))
		code.WriteRune(' ')
		g.generateType(code, c)
		code.WriteRune('\n')
	}
	for _, r := range table.Relationships {
		switch {
		case r.MappedTable != nil:
			switch r.Type {
			case pqt.RelationshipTypeOneToMany,
				pqt.RelationshipTypeOneToManySelfReferencing,
				pqt.RelationshipTypeManyToMany,
				pqt.RelationshipTypeManyToManySelfReferencing:
				if r.MappedBy != "" {
					code.WriteString(g.public(r.MappedBy))
				} else {
					code.WriteString(g.public(r.MappedTable.Name) + "s")
				}
				code.WriteRune(' ')
				fmt.Fprintf(code, "[]*%sEntity", g.private(r.MappedTable.Name))
				code.WriteRune('\n')
			case pqt.RelationshipTypeOneToOneBidirectional,
				pqt.RelationshipTypeOneToOneUnidirectional, // TODO: remove?
				pqt.RelationshipTypeOneToOneSelfReferencing:
				if r.MappedBy != "" {
					code.WriteString(g.public(r.MappedBy))
				} else {
					code.WriteString(g.public(r.MappedTable.Name))
				}
				code.WriteRune(' ')
				fmt.Fprintf(code, "*%sEntity", g.private(r.MappedTable.Name))
				code.WriteRune('\n')
			}
		case r.InversedTable != nil:
			switch r.Type {
			case pqt.RelationshipTypeManyToMany,
				pqt.RelationshipTypeManyToManySelfReferencing:
				if r.InversedBy != "" {
					code.WriteString(g.public(r.InversedBy))
				} else {
					code.WriteString(g.public(r.InversedTable.Name) + "s")
				}
				code.WriteRune(' ')
				fmt.Fprintf(code, "[]*%sEntity", g.private(r.InversedTable.Name))
				code.WriteRune('\n')
			case pqt.RelationshipTypeOneToMany,
				pqt.RelationshipTypeOneToManySelfReferencing,
				pqt.RelationshipTypeOneToOneBidirectional,
				pqt.RelationshipTypeOneToOneUnidirectional,
				pqt.RelationshipTypeOneToOneSelfReferencing:
				if r.InversedBy != "" {
					code.WriteString(g.public(r.InversedBy))
				} else {
					code.WriteString(g.public(r.InversedTable.Name))
				}
				code.WriteRune(' ')
				fmt.Fprintf(code, "*%sEntity", g.private(r.InversedTable.Name))
				code.WriteRune('\n')
			}
		}
	}
	code.WriteString("}\n")
}

func (g *generator) generateType(code *bytes.Buffer, c *pqt.Column) {
	var t string

	if str, ok := c.Type.(fmt.Stringer); ok {
		t = str.String()
	} else {
		t = "struct{}"
	}

	mandatory := c.NotNull || c.PrimaryKey

	switch tp := c.Type.(type) {
	case pqt.MappableType:
		for _, mapto := range tp.Mapping {
			switch mt := mapto.(type) {
			case BuiltinType:
				t = generateBuiltinType(mt, mandatory)
			case CustomType:
				t = mt.String()
			}
		}
	case BuiltinType:
		t = generateBuiltinType(tp, mandatory)
	case pqt.BaseType:
		t = generateBaseType(tp, mandatory)
	}

	code.WriteString(t)
}

func (g *generator) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	g.generateConstantsColumns(code, table)
	g.generateConstantsConstraints(code, table)
	code.WriteString(")\n")
}

func (g *generator) generateConstantsColumns(code *bytes.Buffer, table *pqt.Table) {
	fmt.Fprintf(code, `table%s = "%s"`, g.public(table.Name), table.FullName())
	code.WriteRune('\n')

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(code, `table%sColumn%s = "%s"`, g.public(table.Name), g.public(name), name)
		code.WriteRune('\n')
	}
}

func (g *generator) generateConstantsConstraints(code *bytes.Buffer, table *pqt.Table) {
	for _, c := range tableConstraints(table) {
		name := fmt.Sprintf("%s_%s", c.Table.Name, pqt.JoinColumns(c.Columns, "_"))
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			fmt.Fprintf(code, `table%sConstraint%sCheck = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypePrimaryKey:
			fmt.Fprintf(code, `table%sConstraintPrimaryKey = "%s"`, g.public(table.Name), c.String())
		case pqt.ConstraintTypeForeignKey:
			fmt.Fprintf(code, `table%sConstraint%sForeignKey = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeExclusion:
			fmt.Fprintf(code, `table%sConstraint%sExclusion = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeUnique:
			fmt.Fprintf(code, `table%sConstraint%sUnique = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeIndex:
			fmt.Fprintf(code, `table%sConstraint%sIndex = "%s"`, g.public(table.Name), g.public(name), c.String())
		}

		code.WriteRune('\n')
	}
}

func (g *generator) generateColumns(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("var (\n")
	code.WriteRune('\n')

	code.WriteString("table")
	code.WriteString(g.public(table.Name))
	code.WriteString("Columns = []string{\n")

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(code, "table%sColumn%s", g.public(table.Name), g.public(name))
		code.WriteRune(',')
		code.WriteRune('\n')
	}
	code.WriteString("}")
	code.WriteString(")\n")
}

func (g *generator) generateQueries(code *bytes.Buffer, table *pqt.Table) {

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

func (g *generator) private(s string) string {
	return snake(s, true, g.acronyms)
}

func (g *generator) public(s string) string {
	return snake(s, false, g.acronyms)
}

func nullable(typeA, typeB string, mandatory bool) string {
	if mandatory {
		return typeA
	}
	return typeB
}

func tableConstraints(t *pqt.Table) []*pqt.Constraint {
	var constraints []*pqt.Constraint
	for _, c := range t.Columns {
		constraints = append(constraints, c.Constraints()...)
	}

	return append(constraints, t.Constraints...)
}

func generateBaseType(t pqt.Type, mandatory bool) string {
	switch t {
	case pqt.TypeText():
		return nullable("string", "nilt.String", mandatory)
	case pqt.TypeBool():
		return nullable("bool", "nilt.Bool", mandatory)
	case pqt.TypeIntegerSmall():
		return "int16"
	case pqt.TypeInteger():
		return nullable("int32", "nilt.Int32", mandatory)
	case pqt.TypeIntegerBig():
		return nullable("int64", "nilt.Int64", mandatory)
	case pqt.TypeSerial():
		return "int32"
	case pqt.TypeSerialSmall():
		return "int16"
	case pqt.TypeSerialBig():
		return "int64"
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		return nullable("time.Time", "*time.Time", mandatory)
	case pqt.TypeReal():
		return nullable("float32", "nilt.Float32", mandatory)
	case pqt.TypeDoublePrecision():
		return nullable("float64", "nilt.Float64", mandatory)
	default:
		gt := t.String()
		switch {
		case strings.HasPrefix(gt, "SMALLINT["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "INTEGER["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "BIGINT["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "TEXT["):
			return "pqt.ArrayString"
		case strings.HasPrefix(gt, "DECIMAL"):
			return nullable("float32", "nilt.Float32", mandatory)
		case strings.HasPrefix(gt, "VARCHAR"):
			return nullable("string", "nilt.String", mandatory)
		default:
			return "struct{}"
		}
	}
}

func generateBuiltinType(t BuiltinType, mandatory bool) (r string) {
	switch types.BasicKind(t) {
	case types.Bool:
		r = nullable("bool", "nilt.Bool", mandatory)
	case types.Int:
		r = nullable("int", "nilt.Int", mandatory)
	case types.Int8:
		r = nullable("int8", "*int8", mandatory)
	case types.Int16:
		r = nullable("int16", "*int16", mandatory)
	case types.Int32:
		r = nullable("int32", "nilt.Int32", mandatory)
	case types.Int64:
		r = nullable("int64", "nilt.Int64", mandatory)
	case types.Uint:
		r = nullable("uint", "*uint", mandatory)
	case types.Uint8:
		r = nullable("uint8", "*uint8", mandatory)
	case types.Uint16:
		r = nullable("uint16", "*uint16", mandatory)
	case types.Uint32:
		r = nullable("uint32", "nilt.Uint32", mandatory)
	case types.Uint64:
		r = nullable("uint64", "*uint64", mandatory)
	case types.Float32:
		r = nullable("float32", "nilt.Float32", mandatory)
	case types.Float64:
		r = nullable("float64", "nilt.Float64", mandatory)
	case types.Complex64:
		r = nullable("complex64", "*complex64", mandatory)
	case types.Complex128:
		r = nullable("complex128", "*complex128", mandatory)
	case types.String:
		r = nullable("string", "nilt.String", mandatory)
	default:
		r = "invalid"
	}

	return
}
