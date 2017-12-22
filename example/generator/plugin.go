package main

import (
	"fmt"

	"github.com/piotrkowalczuk/pqt"
)

type generator struct{}

func (g *generator) PropertyType(c *pqt.Column, m int32) string {
	if c.Type == pqt.TypeSerialBig() {
		if m == 1 {
			return "int64"
		}
	}
	return ""
}

func (g *generator) WhereClause(c *pqt.Column) string {
	if c.Type == pqt.TypeSerialBig() {
		return fmt.Sprintf(`
// %s is an empty struct, ignore
`, c.Name)
	}
	return ""
}

func (g *generator) ScanClause(_ *pqt.Column) string {
	return ""
}
func (g *generator) SetClause(_ *pqt.Column) string {
	return ""
}

func (g *generator) Static(_ *pqt.Schema) string {
	return ""
}
