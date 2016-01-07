package pqtgo

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/piotrkowalczuk/pqcnstr"
	"github.com/piotrkowalczuk/pqt"
)

type generator struct {
	acronyms map[string]string
}

func Generator() *generator {
	return &generator{}
}

func (g *generator) Generate(s *pqt.Schema) ([]byte, error) {
	code, err := g.generate(s)
	if err != nil {
		return nil, err
	}

	return code.Bytes(), nil
}

func (g *generator) GenerateTo(s *pqt.Schema, w io.Writer) error {
	code, err := g.generate(s)
	if err != nil {
		return err
	}

	_, err = code.WriteTo(w)
	return err
}

func (g *generator) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	code := bytes.NewBufferString("package main\n")

	for _, table := range s.Tables {
		g.generateConstants(code, table)
		g.generateConstraints(code, table)
		g.generateColumns(code, table)
		g.generateEntity(code, table)
	}

	return code, nil
}

func (g *generator) SetAcronyms(acronyms map[string]string) *generator {
	g.acronyms = acronyms

	return g
}

func (g *generator) generateEntity(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("type " + g.private(table.Name) + "Entity struct {")
	for _, c := range table.Columns {
		code.WriteString(g.public(c.Name))
		code.WriteRune(' ')
		g.generateType(code, c)
		code.WriteRune('\n')
	}
	code.WriteString("}\n")
}

func (g *generator) generateType(code *bytes.Buffer, c *pqt.Column) {
	t := "struct{}"
	mandatory := c.NotNull || c.PrimaryKey

	switch c.Type {
	case pqt.TypeText():
		t = nullable("string", "nilt.String", mandatory)
	case pqt.TypeBool():
		t = nullable("bool", "nilt.Bool", mandatory)
	case pqt.TypeIntegerSmall():
		t = "int16"
	case pqt.TypeInteger():
		t = nullable("int32", "nilt.Int32", mandatory)
	case pqt.TypeIntegerBig():
		t = nullable("int64", "nilt.Int64", mandatory)
	case pqt.TypeSerial():
		t = "uint32"
	case pqt.TypeSerialSmall():
		t = "uint16"
	case pqt.TypeSerialBig():
		t = "uint64"
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		t = nullable("time.Time", "*time.Time", mandatory)
	case pqt.TypeMoney(), pqt.TypeReal():
		t = nullable("float32", "nilt.Float32", mandatory)
	case pqt.TypeDoublePrecision():
		t = nullable("float64", "nilt.Float64", mandatory)
	}

	gt := c.Type.String()
	switch {
	case strings.HasPrefix(gt, "DECIMAL"):
		t = nullable("float32", "nilt.Float32", mandatory)
	case strings.HasPrefix(gt, "VARCHAR"):
		t = nullable("float64", "nilt.Float64", mandatory)
	}

	code.WriteString(t)
}

func (g *generator) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	if table.Schema != nil {
		fmt.Fprintf(code, `table%s = "%s.%s"`, g.public(table.Name), table.Schema.Name, table.Name)
	} else {
		fmt.Fprintf(code, `table%s = "%s"`, g.public(table.Name), table.Name)
	}
	code.WriteRune('\n')

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(code, `table%sColumn%s = "%s"`, g.public(table.Name), g.public(name), name)
		code.WriteRune('\n')
	}
	code.WriteString(")\n")

}

func (g *generator) generateConstraints(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	for _, c := range tableConstraints(table) {
		switch c.Type {
		case pqcnstr.KindCheck:
			fmt.Fprintf(code, `table%sConstraint%sCheck = "%s"`, g.public(table.Name), g.public(c.Name()), c.String())
		case pqcnstr.KindPrimaryKey:
			fmt.Fprintf(code, `table%sConstraintPrimaryKey = "%s"`, g.public(table.Name), c.String())
		case pqcnstr.KindForeignKey:
			fmt.Fprintf(code, `table%sConstraint%sForeignKey = "%s"`, g.public(table.Name), g.public(c.Name()), c.String())
		case pqcnstr.KindExclusion:
			fmt.Fprintf(code, `table%sConstraint%sExclusion = "%s"`, g.public(table.Name), g.public(c.Name()), c.String())
		case pqcnstr.KindUnique:
			fmt.Fprintf(code, `table%sConstraint%sUnique = "%s"`, g.public(table.Name), g.public(c.Name()), c.String())
		case pqcnstr.KindIndex:
			fmt.Fprintf(code, `table%sConstraint%sIndex = "%s"`, g.public(table.Name), g.public(c.Name()), c.String())
		}

		code.WriteRune('\n')
	}
	code.WriteString(")\n")
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
	constraints := make([]*pqt.Constraint, 0)
	for _, c := range t.Columns {
		if cnstr, ok := c.Constraint(); ok {
			constraints = append(constraints, cnstr)
		}
	}

	return append(constraints, t.Constraints...)
}
