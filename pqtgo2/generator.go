package pqtgo2

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/piotrkowalczuk/pqt"
)

var (
	mapping = map[string]struct {
		given     string
		mandatory string
		optional  string
		criteria  string
	}{}
)

// Generator ...
type Generator struct {
	acronyms map[string]string
	imports  []string
	pkg      string
}

// NewGenerator ...
func NewGenerator() *Generator {
	return &Generator{
		pkg: "main",
	}
}

// SetAcronyms ...
func (g *Generator) SetAcronyms(acronyms map[string]string) *Generator {
	g.acronyms = acronyms

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

func (g *Generator) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	code := bytes.NewBuffer(nil)

	g.generatePackage(code)
	g.generateImports(code, s)
	for _, table := range s.Tables {
		g.generateConstants(code, table)
		g.generateColumns(code, table)
		g.generateEntity(code, table)
		g.generateCriteria(code, table)
		g.generateRepository(code, table)
	}

	return code, nil
}

func (g *Generator) generatePackage(w io.Writer) {
	fmt.Fprintf(w, "package %s \n", g.pkg)
}

func (g *Generator) generateImports(w io.Writer, s *pqt.Schema) {
	imports := []string{}

	for _, t := range s.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(CustomType); ok {
				imports = append(imports, ct.mandatoryTypeOf.PkgPath())
				imports = append(imports, ct.criteriaTypeOf.PkgPath())
				imports = append(imports, ct.optionalTypeOf.PkgPath())
			}
		}
	}

	fmt.Fprintln(w, "import (")
	for _, i := range imports {
		fmt.Fprintf(w, `"%s"`, i)
	}
	fmt.Fprintln(w, ")")
}

func (g *Generator) generateConstants(w io.Writer, t *pqt.Table) {
	// TODO: implement
}
func (g *Generator) generateEntity(w io.Writer, t *pqt.Table) {
	fmt.Fprintln(w, "type %sEntity struct {", g.private(t.Name))

	fmt.Fprintln(w, "}")
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

func (g *Generator) private(s string) string {
	return snake(s, true, g.acronyms)
}

func (g *Generator) public(s string) string {
	return snake(s, false, g.acronyms)
}
