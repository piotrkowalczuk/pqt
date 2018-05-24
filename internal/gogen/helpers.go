package gogen

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

type structField struct {
	Name     string
	Type     string
	Tags     reflect.StructTag
	ReadOnly bool
}

func closeBrace(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, `
		}`)
	}
}

func columnMode(c *pqt.Column, m int32) int32 {
	switch m {
	case pqtgo.ModeCriteria:
	case pqtgo.ModeMandatory:
	case pqtgo.ModeOptional:
	default:
		if c.NotNull || c.PrimaryKey {
			m = pqtgo.ModeMandatory
		}
	}
	return m
}

func columnForeignName(c *pqt.Column) string {
	return c.Table.Name + "_" + c.Name
}

func or(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
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

func uniqueConstraints(t *pqt.Table) []*pqt.Constraint {
	var unique []*pqt.Constraint
	for _, c := range t.Constraints {
		if c.Type == pqt.ConstraintTypeUnique || c.Type == pqt.ConstraintTypeUniqueIndex {
			unique = append(unique, c)
		}
	}
	if len(unique) < 1 {
		return nil
	}
	return unique
}

func sqlSelector(c *pqt.Column, id string) string {
	if !c.IsDynamic {
		return pqtfmt.Public("table", c.Table.Name, "column", c.Name)
	}
	sel := c.Func.Name
	sel += "("
	for i := range c.Func.Args {
		if i != 0 {
			sel += ", "
		}
		sel += "t%d."
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

func functionName(f *pqt.Function) string {
	if f.Name == "" {
		fmt.Println(fmt.Sprintf("function name is missing\n\n%s", debug.Stack()))
		os.Exit(1)
	}
	return f.Name
}
