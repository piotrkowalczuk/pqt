package gogen

import (
	"fmt"
	"html/template"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) Entity(t *pqt.Table) {
	g.Printf(`
// %sEntity ...`, pqtfmt.Public(t.Name))
	g.Printf(`
type %sEntity struct{`, pqtfmt.Public(t.Name))
	for prop := range g.entityPropertiesGenerator(t) {
		g.Printf(`
// %s ...`, pqtfmt.Public(prop.Name))
		if prop.ReadOnly {
			g.Printf(`
// %s is read only`, pqtfmt.Public(prop.Name))
		}
		if prop.Tags != "" {
			g.Printf(`
%s %s %s`, pqtfmt.Public(prop.Name), prop.Type, prop.Tags)
		} else {
			g.Printf(`
%s %s`,
				pqtfmt.Public(prop.Name),
				prop.Type,
			)
		}
	}
	g.Print(`}`)
}

func (g *Generator) EntityProp(t *pqt.Table) {
	g.Printf(`
		func (e *%sEntity) %s(cn string) (interface{}, bool) {`, pqtfmt.Public(t.Name), pqtfmt.Public("prop"))
	g.Println(`
		switch cn {`)

ColumnsLoop:
	for _, c := range t.Columns {
		g.Printf(`
			case %s:`, pqtfmt.Public("table", t.Name, "column", c.Name))
		for _, plugin := range g.Plugins {
			if txt := plugin.ScanClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(fmt.Sprintf(`
					return %s, true`, txt))
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(g, map[string]interface{}{
					"selector": fmt.Sprintf("e.%s", pqtfmt.Public(c.Name)),
				}); err != nil {
					panic(err)
				}
				g.Println("")
				continue ColumnsLoop
			}
		}
		switch {
		case g.isArray(c, pqtgo.ModeDefault):
			pn := pqtfmt.Public(c.Name)
			switch g.columnType(c, pqtgo.ModeDefault) {
			case "pq.Int64Array":
				g.Printf(`if e.%s == nil { e.%s = []int64{} }`, pn, pn)
			case "pq.StringArray":
				g.Printf(`if e.%s == nil { e.%s = []string{} }`, pn, pn)
			case "pq.Float64Array":
				g.Printf(`if e.%s == nil { e.%s = []float64{} }`, pn, pn)
			case "pq.BoolArray":
				g.Printf(`if e.%s == nil { e.%s = []bool{} }`, pn, pn)
			case "pq.ByteaArray":
				g.Printf(`if e.%s == nil { e.%s = [][]byte{} }`, pn, pn)
			}

			g.Printf(`
				return &e.%s, true`, pqtfmt.Public(c.Name))
		case g.canBeNil(c, pqtgo.ModeDefault):
			g.Printf(`
				return e.%s, true`,
				pqtfmt.Public(c.Name),
			)
		default:
			g.Printf(`
				return &e.%s, true`,
				pqtfmt.Public(c.Name),
			)
		}
	}

	g.Print(`
	default:
		return nil, false
	}
}`)
}

func (g *Generator) EntityProps(t *pqt.Table) {
	g.Printf(`
		func (e *%sEntity) %s(cns ...string) ([]interface{}, error) {`, pqtfmt.Public(t.Name), pqtfmt.Public("props"))
	g.Printf(`
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
		pqtfmt.Public("table", t.Name, "columns"),
		pqtfmt.Public("prop"),
	)
	g.Print(`
		}`)
}
