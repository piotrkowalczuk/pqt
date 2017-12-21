package gogen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

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
	return formatter.Type(c.Type, m)
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
