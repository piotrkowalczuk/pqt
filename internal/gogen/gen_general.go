package gogen

import (
	"fmt"
	"text/template"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/print"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
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
	tableName := pqtfmt.Public(t.Name)

	g.Printf(`
type %sCriteria struct {`, tableName)
	for _, c := range t.Columns {
		if t := g.columnType(c, pqtgo.ModeCriteria); t != "<nil>" {
			g.Printf(`
%s %s`, pqtfmt.Public(c.Name), t)
		}
	}
	g.Printf(`
	operator string
	child, sibling, parent *%sCriteria
}`, tableName)
}

func (g *Generator) Errors() {
	g.Printf(`
// RetryTransaction can be returned by user defined function when a transaction is rolled back and logic repeated.
var RetryTransaction = errors.New("retry transaction")`)
}

func (g *Generator) Operand(t *pqt.Table) {
	tableName := pqtfmt.Public(t.Name)

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
%s = "%s"`, pqtfmt.Public("table", t.Name), t.FullName())

	for _, c := range t.Columns {
		g.Printf(`
%s = "%s"`, pqtfmt.Public("table", t.Name, "column", c.Name), c.Name)
	}

	g.Printf(`
)

var %s = []string{`, pqtfmt.Public("table", t.Name, "columns"))

	for _, c := range t.Columns {
		g.Printf(`
%s,`, pqtfmt.Public("table", t.Name, "column", c.Name))
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
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraint", name, "Check"), c.String())
		case pqt.ConstraintTypePrimaryKey:
			g.Printf(`
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraintPrimaryKey"), c.String())
		case pqt.ConstraintTypeForeignKey:
			g.Printf(`
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraint", name, "ForeignKey"), c.String())
		case pqt.ConstraintTypeExclusion:
			g.Printf(`
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraint", name, "Exclusion"), c.String())
		case pqt.ConstraintTypeUnique:
			g.Printf(`
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraint", name, "Unique"), c.String())
		case pqt.ConstraintTypeIndex:
			g.Printf(`
%s = "%s"`, pqtfmt.Public("table", c.PrimaryTable.Name, "constraint", name, "Index"), c.String())
		}
	}
	g.Printf(`
)`)
}

func (g *Generator) FindExpr(t *pqt.Table) {
	g.Printf(`
type %sFindExpr struct {`, pqtfmt.Public(t.Name))
	g.Printf(`
%s *%sCriteria`, pqtfmt.Public("where"), pqtfmt.Public(t.Name))
	g.Printf(`
%s, %s int64`, pqtfmt.Public("offset"), pqtfmt.Public("limit"))
	g.Printf(`
%s []string`, pqtfmt.Public("columns"))
	g.Printf(`
%s []RowOrder`, pqtfmt.Public("orderBy"))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
%s *%sJoin`, pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)), pqtfmt.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) CountExpr(t *pqt.Table) {
	g.Printf(`
type %sCountExpr struct {`, pqtfmt.Public(t.Name))
	g.Printf(`
%s *%sCriteria`, pqtfmt.Public("where"), pqtfmt.Public(t.Name))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
%s *%sJoin`, pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)), pqtfmt.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) Join(t *pqt.Table) {
	g.Printf(`
type %sJoin struct {`, pqtfmt.Public(t.Name))
	g.Printf(`
%s, %s *%sCriteria`, pqtfmt.Public("on"), pqtfmt.Public("where"), pqtfmt.Public(t.Name))
	g.Printf(`
%s bool`, pqtfmt.Public("fetch"))
	g.Printf(`
%s JoinType`, pqtfmt.Public("kind"))
	for _, r := range joinableRelationships(t) {
		g.Printf(`
Join%s *%sJoin`, pqtfmt.Public(or(r.InversedName, r.InversedTable.Name)), pqtfmt.Public(r.InversedTable.Name))
	}
	g.Print(`
}`)
}

func (g *Generator) Patch(t *pqt.Table) {
	g.Printf(`
type %sPatch struct {`, pqtfmt.Public(t.Name))

ArgumentsLoop:
	for _, c := range t.Columns {
		if c.PrimaryKey {
			continue ArgumentsLoop
		}

		if t := g.columnType(c, pqtgo.ModeOptional); t != "<nil>" {
			g.Printf(`
%s %s`,
				pqtfmt.Public(c.Name),
				t,
			)
		}
	}
	g.Print(`
}`)
}

func (g *Generator) Iterator(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	g.Printf(`
// %sIterator is not thread safe.
type %sIterator struct {
	rows Rows
	cols []string
	expr *%sFindExpr
}`, entityName,
		entityName,
		pqtfmt.Public(t.Name))

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
		pqtfmt.Public(t.Name),
		entityName,
		pqtfmt.Public(t.Name),
		entityName,
		entityName,
		pqtfmt.Public("props"))

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
	name := pqtfmt.Public(t.Name)
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
					"selector": fmt.Sprintf("c.%s", pqtfmt.Public(c.Name)),
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
				if c.%s != nil {`, pqtfmt.Public(c.Name))
		}
		if g.isNullable(c, pqtgo.ModeCriteria) {
			braces++
			g.Printf(`
				if c.%s.Valid {`, pqtfmt.Public(c.Name))
		}
		if g.isType(c, pqtgo.ModeCriteria, "time.Time") {
			braces++
			g.Printf(`
				if !c.%s.IsZero() {`, pqtfmt.Public(c.Name))
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
					pqtfmt.Public("table", c.Columns[i].Table.Name, "column", c.Columns[i].Name),
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
				pqtfmt.Public("table", t.Name, "column", c.Name),
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
			pqtfmt.Public(c.Name),
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
	}`)
}

func (g *Generator) ScanRows(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	funcName := pqtfmt.Public("scan", t.Name, "rows")
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
		g.Printf("&ent.%s,\n", pqtfmt.Public(c.Name))
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

func (g *Generator) Funcs() {
	g.Print(`
	// LogFunc represents function that can be passed into repository to log query result.
	type LogFunc func(err error, ent, fnc, sql string, args ...interface{})`)
}

func (g *Generator) Interfaces() {
	g.Print(`
	// Rows ...
	type Rows interface {
		io.Closer
		ColumnTypes() ([]*sql.ColumnType, error)
		Columns() ([]string, error)
		Err() error
		Next() bool
		NextResultSet() bool
		Scan(dst ...interface{}) error
	}`)
}

func (g *Generator) Statics() {
	code := `
const (
	JoinInner = iota
	JoinLeft
	JoinRight
	JoinCross
	JoinDoNot
)

type JoinType int

func (jt JoinType) String() string {
	switch jt {

	case JoinInner:
		return "INNER JOIN"
	case JoinLeft:
		return "LEFT JOIN"
	case JoinRight:
		return "RIGHT JOIN"
	case JoinCross:
		return "CROSS JOIN"
	default:
		return ""
	}
}

// Actionable returns true if JoinType is one of the known type except JoinDoNot.
func (jt JoinType) Actionable() bool {
	switch jt {
	case JoinInner, JoinLeft, JoinRight, JoinCross:
		return true
	default:
		return false
	}
}

// ErrorConstraint returns the error constraint of err if it was produced by the pq library.
// Otherwise, it returns empty string.
func ErrorConstraint(err error) string {
	if err == nil {
		return ""
	}
	if pqerr, ok := err.(*pq.Error); ok {
		return pqerr.Constraint
	}

	return ""
}

type RowOrder struct {
	Name string
	Descending bool
}

type NullInt64Array struct {
	pq.Int64Array
	Valid  bool
}

func (n *NullInt64Array) Scan(value interface{}) error {
	if value == nil {
		n.Int64Array, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.Int64Array.Scan(value)
}

type NullFloat64Array struct {
	pq.Float64Array
	Valid  bool
}

func (n *NullFloat64Array) Scan(value interface{}) error {
	if value == nil {
		n.Float64Array, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.Float64Array.Scan(value)
}

type NullBoolArray struct {
	pq.BoolArray
	Valid  bool
}

func (n *NullBoolArray) Scan(value interface{}) error {
	if value == nil {
		n.BoolArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.BoolArray.Scan(value)
}

type NullStringArray struct {
	pq.StringArray
	Valid  bool
}

func (n *NullStringArray) Scan(value interface{}) error {
	if value == nil {
		n.StringArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.StringArray.Scan(value)
}

type NullByteaArray struct {
	pq.ByteaArray
	Valid  bool
}

func (n *NullByteaArray) Scan(value interface{}) error {
	if value == nil {
		n.ByteaArray, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	return n.ByteaArray.Scan(value)
}


const (
	jsonArraySeparator     = ","
	jsonArrayBeginningChar = "["
	jsonArrayEndChar       = "]"
)

// JSONArrayInt64 is a slice of int64s that implements necessary interfaces.
type JSONArrayInt64 []int64

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayInt64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayInt64, 0)
		}
		return nil
	}

	var tmp []string
	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
	}

	if l == 2 {
		*a = make(JSONArrayInt64, 0)
		return nil
	}

	if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
		return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
	*a = make(JSONArrayInt64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s at index %d", v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayInt64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatInt(v, 10)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// JSONArrayString is a slice of strings that implements necessary interfaces.
type JSONArrayString []string

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayString) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayString, 0)
		}
		return nil
	}

	switch t := src.(type) {
	case []byte:
		return json.Unmarshal(t, a)
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayString) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// JSONArrayFloat64 is a slice of int64s that implements necessary interfaces.
type JSONArrayFloat64 []float64

// Scan satisfy sql.Scanner interface.
func (a *JSONArrayFloat64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(JSONArrayFloat64, 0)
		}
		return nil
	}

	var (
		tmp []string
		srcs string
	)

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
	}

	if l == 2 {
		*a = make(JSONArrayFloat64, 0)
		return nil
	}

	if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
		return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
	*a = make(JSONArrayFloat64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s at index %d", v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a JSONArrayFloat64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}


var (
	// Space is a shorthand composition option that holds space.
	Space = &CompositionOpts{
		Joint: " ",
	}
	// And is a shorthand composition option that holds AND operator.
	And = &CompositionOpts{
		Joint: " AND ",
	}
	// Or is a shorthand composition option that holds OR operator.
	Or = &CompositionOpts{
		Joint: " OR ",
	}
	// Comma is a shorthand composition option that holds comma.
	Comma = &CompositionOpts{
		Joint: ", ",
	}
)

// CompositionOpts is a container for modification that can be applied.
type CompositionOpts struct {
	Joint                         string
	PlaceholderFuncs, SelectorFuncs []string
	PlaceholderCast, SelectorCast   string
	IsJSON                        bool
	IsDynamic                     bool
}

// CompositionWriter is a simple wrapper for WriteComposition function.
type CompositionWriter interface {
	// WriteComposition is a function that allow custom struct type to be used as a part of criteria.
	// It gives possibility to write custom query based on object that implements this interface.
	WriteComposition(string, *Composer, *CompositionOpts) error
}

// Composer holds buffer, arguments and placeholders count.
// In combination with external buffet can be also used to also generate sub-queries.
// To do that simply write buffer to the parent buffer, composer will hold all arguments and remember number of last placeholder.
type Composer struct {
	buf     bytes.Buffer
	args    []interface{}
	counter int
	Dirty   bool
}

// NewComposer allocates new Composer with inner slice of arguments of given size.
func NewComposer(size int64) *Composer {
	return &Composer{
		counter: 1,
		args:    make([]interface{}, 0, size),
	}
}

// WriteString appends the contents of s to the query buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with bytes ErrTooLarge.
func (c *Composer) WriteString(s string) (int, error) {
	return c.buf.WriteString(s)
}

// Write implements io Writer interface.
func (c *Composer) Write(b []byte) (int, error) {
	return c.buf.Write(b)
}

// Read implements io Reader interface.
func (c *Composer) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}

// ResetBuf resets internal buffer.
func (c *Composer) ResetBuf() {
	c.buf.Reset()
}

// String implements fmt Stringer interface.
func (c *Composer) String() string {
	return c.buf.String()
}

// WritePlaceholder writes appropriate placeholder to the query buffer based on current state of the composer.
func (c *Composer) WritePlaceholder() error {
	if _, err := c.buf.WriteString("$"); err != nil {
		return err
	}
	if _, err := c.buf.WriteString(strconv.Itoa(c.counter)); err != nil {
		return err
	}

	c.counter++
	return nil
}

func (c *Composer) WriteAlias(i int) error {
	if i < 0 {
		return nil
	}
	if _, err := c.buf.WriteString("t"); err != nil {
		return err
	}
	if _, err := c.buf.WriteString(strconv.Itoa(i)); err != nil {
		return err
	}
	if _, err := c.buf.WriteString("."); err != nil {
		return err
	}
	return nil
}

// Len returns number of arguments.
func (c *Composer) Len() int {
	return c.counter
}

// Add appends list with new element.
func (c *Composer) Add(arg interface{}) {
	c.args = append(c.args, arg)
}

// Args returns all arguments stored as a slice.
func (c *Composer) Args() []interface{} {
	return c.args
}`

	g.Print(code)
}

func (g *Generator) PluginsStatics(s *pqt.Schema) {
	for _, plugin := range g.Plugins {
		if txt := plugin.Static(s); txt != "" {
			g.Print(txt)
			g.Print("\n\n")
		}
	}
}
