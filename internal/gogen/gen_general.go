package gogen

import (
	"fmt"
	"text/template"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
	"github.com/piotrkowalczuk/pqt/internal/print"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

type Generator struct {
	print.Printer
	Plugins []Plugin
	Version float64
}

// Package generates package header.
func (g *Generator) Package(pkg string) {
	if pkg == "" {
		pkg = "main"
	}
	g.Printf("package %s\n", pkg)
}

func (g *Generator) Imports(s *pqt.Schema, fixed ...string) {
	imports := []string{
		"github.com/m4rw3r/uuid",
	}
	imports = append(imports, fixed...)

	appendIfNotEmpty := func(slice []string, elem string) []string {
		if elem != "" {
			return append(slice, elem)
		}
		return slice
	}
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(pqtgo.CustomType); ok {
				imports = appendIfNotEmpty(imports, ct.TypeOf(pqtgo.ModeMandatory).PkgPath())
				imports = appendIfNotEmpty(imports, ct.TypeOf(pqtgo.ModeOptional).PkgPath())
				imports = appendIfNotEmpty(imports, ct.TypeOf(pqtgo.ModeCriteria).PkgPath())
			}
		}
	}

	g.Println("import(")
	for _, imp := range imports {
		g.Print(`"`)
		g.Print(imp)
		g.Println(`"`)
	}
	g.Println(")")
}

func (g *Generator) Criteria(t *pqt.Table) {
	tableName := formatter.Public(t.Name)

	g.Printf(`
type %sCriteria struct {`, tableName)
	for _, c := range t.Columns {
		if t := g.columnType(c, pqtgo.ModeCriteria); t != "<nil>" {
			g.Printf(`
%s %s`, formatter.Public(c.Name), t)
		}
	}
	g.Printf(`
	operator string
	child, sibling, parent *%sCriteria
}`, tableName)
}

func (g *Generator) Operand(t *pqt.Table) {
	tableName := formatter.Public(t.Name)

	g.Printf(`
func %sOperand(operator string, operands ...*%sCriteria) *%sCriteria {
	if len(operands) == 0 {
		return &%sCriteria{operator: operator}
	}

	parent := &%sCriteria{
		operator: operator,
		child: operands[0],
	}

	for i := 0; i < len(operands); i++ {
		if i < len(operands)-1 {
			operands[i].sibling = operands[i+1]
		}
		operands[i].parent = parent
	}

	return parent
}`, tableName, tableName, tableName, tableName, tableName)
	g.Printf(`

func %sOr(operands ...*%sCriteria) *%sCriteria {
	return %sOperand("OR", operands...)
}`, tableName, tableName, tableName, tableName)
	g.Printf(`

func %sAnd(operands ...*%sCriteria) *%sCriteria {
	return %sOperand("AND", operands...)
}`, tableName, tableName, tableName, tableName)
}

func (g *Generator) Columns(t *pqt.Table) {
	g.Printf(`
const (
%s = "%s"`, formatter.Public("table", t.Name), t.FullName())

	for _, c := range t.Columns {
		g.Printf(`
%s = "%s"`, formatter.Public("table", t.Name, "column", c.Name), c.Name)
	}

	g.Printf(`
)

var %s = []string{`, formatter.Public("table", t.Name, "columns"))

	for _, c := range t.Columns {
		g.Printf(`
%s,`, formatter.Public("table", t.Name, "column", c.Name))
	}
	g.Print(`
}`)

}

func (g *Generator) Constraints(t *pqt.Table) {
	g.Printf(`
const (`)
	for _, c := range t.Constraints {
		name := pqt.JoinColumns(c.PrimaryColumns, "_")
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraint", name, "Check"), c.String())
		case pqt.ConstraintTypePrimaryKey:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraintPrimaryKey"), c.String())
		case pqt.ConstraintTypeForeignKey:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraint", name, "ForeignKey"), c.String())
		case pqt.ConstraintTypeExclusion:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraint", name, "Exclusion"), c.String())
		case pqt.ConstraintTypeUnique:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraint", name, "Unique"), c.String())
		case pqt.ConstraintTypeIndex:
			g.Printf(`
%s = "%s"`, formatter.Public("table", c.PrimaryTable.Name, "constraint", name, "Index"), c.String())
		}
	}
	g.Printf(`
)`)
}

func (g *Generator) Repository(t *pqt.Table) {
	g.Printf(`
type %sRepositoryBase struct {
	%s string
	%s []string
	%s *sql.DB
	%s LogFunc
}`,
		formatter.Public(t.Name),
		formatter.Public("table"),
		formatter.Public("columns"),
		formatter.Public("db"),
		formatter.Public("log"),
	)
}

func (g *Generator) FindExpr(t *pqt.Table) {
	g.Printf(`
type %sFindExpr struct {`, formatter.Public(t.Name))
	g.Printf(`
%s *%sCriteria`, formatter.Public("where"), formatter.Public(t.Name))
	g.Printf(`
%s, %s int64`, formatter.Public("offset"), formatter.Public("limit"))
	g.Printf(`
%s []string`, formatter.Public("columns"))
	g.Printf(`
%s []RowOrder`, formatter.Public("orderBy"))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
%s *%sJoin`, formatter.Public("join", or(r.InversedName, r.InversedTable.Name)), formatter.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) CountExpr(t *pqt.Table) {
	g.Printf(`
type %sCountExpr struct {`, formatter.Public(t.Name))
	g.Printf(`
%s *%sCriteria`, formatter.Public("where"), formatter.Public(t.Name))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
%s *%sJoin`, formatter.Public("join", or(r.InversedName, r.InversedTable.Name)), formatter.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) Join(t *pqt.Table) {
	g.Printf(`
type %sJoin struct {`, formatter.Public(t.Name))
	g.Printf(`
%s, %s *%sCriteria`, formatter.Public("on"), formatter.Public("where"), formatter.Public(t.Name))
	g.Printf(`
%s bool`, formatter.Public("fetch"))
	g.Printf(`
%s JoinType`, formatter.Public("kind"))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
Join%s *%sJoin`, formatter.Public(or(r.InversedName, r.InversedTable.Name)), formatter.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) Patch(t *pqt.Table) {
	g.Printf(`
type %sPatch struct {`, formatter.Public(t.Name))

ArgumentsLoop:
	for _, c := range t.Columns {
		if c.PrimaryKey {
			continue ArgumentsLoop
		}

		if t := g.columnType(c, pqtgo.ModeOptional); t != "<nil>" {
			g.Printf(`
%s %s`,
				formatter.Public(c.Name),
				t,
			)
		}
	}
	g.Print(`
}`)
}

func (g *Generator) Iterator(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	g.Printf(`
// %sIterator is not thread safe.
type %sIterator struct {
	rows Rows
	cols []string
	expr *%sFindExpr
}`, entityName,
		entityName,
		formatter.Public(t.Name))

	g.Printf(`
func (i *%sIterator) Next() bool {
	return i.rows.Next()
}

func (i *%sIterator) Close() error {
	return i.rows.Close()
}

func (i *%sIterator) Err() error {
	return i.rows.Err()
}

// Columns is wrapper around sql.Rows.Columns method, that also cache output inside iterator.
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
	cols, err := i.Columns()
	if err != nil {
		return nil, err
	}

	props, err := ent.%s(cols...)
	if err != nil {
		return nil, err
	}`, entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		entityName,
		formatter.Public(t.Name),
		entityName,
		formatter.Public(t.Name),
		entityName,
		entityName,
		formatter.Public("props"))

	if hasJoinableRelationships(t) {
		g.Print(`
		var prop []interface{}`)
	}
	g.scanJoinableRelationships(t, "i.expr")

	g.Print(`
	if err := i.rows.Scan(props...); err != nil {
		return nil, err
	}
	return &ent, nil
}`)
}

func (g *Generator) WhereClause(t *pqt.Table) {
	name := formatter.Public(t.Name)
	fnName := fmt.Sprintf("%sCriteriaWhereClause", name)
	g.Printf(`
		func %s(comp *Composer, c *%sCriteria, id int) (error) {`, fnName, name)

	g.Printf(`
	if c.child == nil {
		return _%s(comp, c, id)
	}
	node := c
	sibling := false
	for {
		if !sibling {
			if node.child != nil {
				if node.parent != nil {
					comp.WriteString("(")
				}
				node = node.child
				continue
			} else {
				comp.Dirty = false
				comp.WriteString("(")
				if err := _%s(comp, node, id); err != nil {
					return err
				}
				comp.WriteString(")")
			}
		}
		if node.sibling != nil {
			sibling = false
			comp.WriteString(" ")
			comp.WriteString(node.parent.operator)
			comp.WriteString(" ")
			node = node.sibling
			continue
		}
		if node.parent != nil {
			sibling = true
			if node.parent.parent != nil {
				comp.WriteString(")")
			}
			node = node.parent
			continue
		}

		break
	}
	return nil
	}`, fnName, fnName)

	g.Printf(`

		func _%sCriteriaWhereClause(comp *Composer, c *%sCriteria, id int) (error) {`, name, name)
ColumnsLoop:
	for _, c := range t.Columns {
		braces := 0
		for _, plugin := range g.Plugins {
			if txt := plugin.WhereClause(c); txt != "" {
				tmpl, err := template.New("root").Parse(txt)
				if err != nil {
					panic(err)
				}
				if err = tmpl.Execute(g, map[string]interface{}{
					"selector": fmt.Sprintf("c.%s", formatter.Public(c.Name)),
					"column":   sqlSelector(c, "id"),
					"composer": "comp",
					"id":       "id",
				}); err != nil {
					panic(err)
				}
				g.NewLine()
				continue ColumnsLoop
			}
		}
		if g.columnType(c, pqtgo.ModeCriteria) == "<nil>" {
			break ColumnsLoop
		}
		if g.canBeNil(c, pqtgo.ModeCriteria) {
			braces++
			g.Printf(`
				if c.%s != nil {`, formatter.Public(c.Name))
		}
		if g.isNullable(c, pqtgo.ModeCriteria) {
			braces++
			g.Printf(`
				if c.%s.Valid {`, formatter.Public(c.Name))
		}
		if g.isType(c, pqtgo.ModeCriteria, "time.Time") {
			braces++
			g.Printf(`
				if !c.%s.IsZero() {`, formatter.Public(c.Name))
		}

		g.Print(
			`if comp.Dirty {
				comp.WriteString(" AND ")
			}`)

		if c.IsDynamic {
			if len(c.Func.Args) > len(c.Columns) {
				panic(fmt.Sprintf("number of function arguments is greater then number of available columns: %s.%s", c.Table.Name, c.Name))
			}
			g.Printf(`
				if _, err := comp.WriteString("%s"); err != nil {
					return err
				}
				if _, err := comp.WriteString("("); err != nil {
					return err
				}`, functionName(c.Func))
			for i, arg := range c.Func.Args {
				if arg.Type != c.Columns[i].Type {
					fmt.Printf("wrong function (%s) argument type, expected %v but got %v\n", functionName(c.Func), arg.Type, c.Columns[i].Type)
				}
				if i != 0 {
					g.Print(`
					if _, err := comp.WriteString(", "); err != nil {
						return err
					}`)
				}
				g.Printf(`
					if err := comp.WriteAlias(id); err != nil {
						return err
					}
					if _, err := comp.WriteString(%s); err != nil {
						return err
					}`,
					formatter.Public("table", c.Columns[i].Table.Name, "column", c.Columns[i].Name),
				)
			}
			g.Print(`
				if _, err := comp.WriteString(")"); err != nil {
					return err
				}`)
		} else {
			g.Printf(`
				if err := comp.WriteAlias(id); err != nil {
					return err
				}
				if _, err := comp.WriteString(%s); err != nil {
					return err
				}`,
				formatter.Public("table", t.Name, "column", c.Name),
			)
		}

		g.Printf(`
			if _, err := comp.WriteString("="); err != nil {
				return err
			}
			if err := comp.WritePlaceholder(); err != nil {
				return err
			}
			comp.Add(c.%s)
			comp.Dirty=true`,
			formatter.Public(c.Name),
		)
		closeBrace(g, braces)
	}
	g.Print(`
	return nil`)
	closeBrace(g, 1)
}

func (g *Generator) JoinClause() {
	g.Print(`
	func joinClause(comp *Composer, jt JoinType, on string) (ok bool, err error) {
		if jt != JoinDoNot {
			switch jt {
			case JoinInner:
				if _, err = comp.WriteString(" INNER JOIN "); err != nil {
					return
				}
			case JoinLeft:
				if _, err = comp.WriteString(" LEFT JOIN "); err != nil {
					return
				}
			case JoinRight:
				if _, err = comp.WriteString(" RIGHT JOIN "); err != nil {
					return
				}
			case JoinCross:
				if _, err = comp.WriteString(" CROSS JOIN "); err != nil {
					return
				}
			default:
				return
			}
			if _, err = comp.WriteString(on); err != nil {
				return
			}
			comp.Dirty = true
			ok = true
			return
		}
		return
	}`,
	)
}

func (g *Generator) ScanRows(t *pqt.Table) {
	entityName := formatter.Public(t.Name)
	funcName := formatter.Public("scan", t.Name, "rows")
	g.Printf(`
		// %s helps to scan rows straight to the slice of entities.
		func %s(rows Rows) (entities []*%sEntity, err error) {`, funcName, funcName, entityName)
	g.Printf(`
		for rows.Next() {
			var ent %sEntity
			err = rows.Scan(
			`, entityName,
	)
	for _, c := range t.Columns {
		g.Printf("&ent.%s,\n", formatter.Public(c.Name))
	}
	g.Print(`)
			if err != nil {
				return
			}

			entities = append(entities, &ent)
		}
		if err = rows.Err(); err != nil {
			return
		}

		return
	}`)
}
