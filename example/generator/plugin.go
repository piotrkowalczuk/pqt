package main

import (
	"fmt"

	"github.com/piotrkowalczuk/pqt"
)

type generator struct {
}

func (g *generator) PropertyType(c *pqt.Column, m int32) (string) {
	if m == 1 {
		if c.Type == pqt.TypeSerialBig() {
			return "int64"
		}
	}
	return ""
}

func (g *generator) WhereClause(c *pqt.Column) (string) {
	if c.Type == pqt.TypeSerialBig() {
		return fmt.Sprintf("// %s is an empty struct, ignore\n", c.Name)
	}
	return ""
}
