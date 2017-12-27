package pqtgogen

import (
	"go/format"
	"io"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
)

type Generator struct {
	Formatter  *Formatter
	Version    float64
	Pkg        string
	Imports    []string
	Plugins    []Plugin
	Components Component

	g *gogen.Generator
	p *print.Printer
}

// Generate ...
func (g *Generator) Generate(s *pqt.Schema) ([]byte, error) {
	if err := g.generate(s); err != nil {
		return nil, err
	}

	return format.Source(g.p.Bytes())
}

// GenerateTo ...
func (g *Generator) GenerateTo(w io.Writer, s *pqt.Schema) error {
	if err := g.generate(s); err != nil {
		return err
	}

	buf, err := format.Source(g.p.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func (g *Generator) generate(s *pqt.Schema) error {
	g.g = &gogen.Generator{
		Version: g.Version,
	}
	for _, p := range g.Plugins {
		g.g.Plugins = append(g.g.Plugins, p)
	}
	g.p = &g.g.Printer

	g.generatePackage()
	g.generateImports(s)
	if g.Components&ComponentRepository != 0 {
		g.generateLogFunc(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 || g.Components&ComponentHelpers != 0 {
		g.generateInterfaces(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
		g.generateJoinClause()
	}
	for _, t := range s.Tables {
		g.generateConstantsAndVariables(t)
		g.generateEntity(t)
		g.generateEntityProp(t)
		g.generateEntityProps(t)
		if g.Components&ComponentHelpers != 0 {
			g.generateScanRows(t)
		}
		if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
			g.generateIterator(t)
			g.generateCriteria(t)
			g.generateFindExpr(t)
			g.generateJoin(t)
		}
		if g.Components&ComponentCount != 0 {
			g.generateCountExpr(t)
		}
		if g.Components&ComponentUpdate != 0 || g.Components&ComponentUpsert != 0 {
			g.generatePatch(t)
		}
		if g.Components&ComponentRepository != 0 {
			g.generateRepository(t)
			if g.Components&ComponentInsert != 0 {
				g.generateRepositoryInsertQuery(t)
				g.generateRepositoryInsert(t)
			}
			if g.Components&ComponentFind != 0 {
				g.generateWhereClause(t)
				g.generateRepositoryFindQuery(t)
				g.generateRepositoryFind(t)
				g.generateRepositoryFindIter(t)
				g.generateRepositoryFindOneByPrimaryKey(t)
				g.generateRepositoryFindOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpdate != 0 {
				g.generateRepositoryUpdateOneByPrimaryKeyQuery(t)
				g.generateRepositoryUpdateOneByPrimaryKey(t)
				g.generateRepositoryUpdateOneByUniqueConstraintQuery(t)
				g.generateRepositoryUpdateOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpsert != 0 {
				g.generateRepositoryUpsertQuery(t)
				g.generateRepositoryUpsert(t)
			}
			if g.Components&ComponentCount != 0 {
				g.generateRepositoryCount(t)
			}
			if g.Components&ComponentDelete != 0 {
				g.generateRepositoryDeleteOneByPrimaryKey(t)
			}
		}
	}
	g.generateStatics(s)

	return g.p.Err
}

func (g *Generator) generatePackage() {
	g.g.Package(g.Pkg)
}

func (g *Generator) generateImports(s *pqt.Schema) {
	g.g.Imports(s, "github.com/m4rw3r/uuid")
}

func (g *Generator) generateEntity(t *pqt.Table) {
	g.g.Entity(t)
	g.g.NewLine()
}

func (g *Generator) generateFindExpr(t *pqt.Table) {
	g.g.FindExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCountExpr(t *pqt.Table) {
	g.g.CountExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCriteria(t *pqt.Table) {
	g.g.Criteria(t)
	g.g.NewLine()
	g.g.Operand(t)
	g.g.NewLine()
}

func (g *Generator) generateJoin(t *pqt.Table) {
	g.g.Join(t)
	g.g.NewLine()
}

func (g *Generator) generatePatch(t *pqt.Table) {
	g.g.Patch(t)
	g.g.NewLine()
}

func (g *Generator) generateIterator(t *pqt.Table) {
	g.g.Iterator(t)
	g.g.NewLine()
}

func (g *Generator) generateRepository(t *pqt.Table) {
	g.g.Repository(t)
	g.g.NewLine()
}

func (g *Generator) generateConstantsAndVariables(t *pqt.Table) {
	g.g.Constraints(t)
	g.g.NewLine()
	g.g.Columns(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsertQuery(t *pqt.Table) {
	g.g.RepositoryInsertQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsert(t *pqt.Table) {
	g.g.RepositoryInsert(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKeyQuery(t *pqt.Table) {
	g.g.RepositoryUpdateOneByPrimaryKeyQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryUpdateOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraintQuery(t *pqt.Table) {
	g.g.RepositoryUpdateOneByUniqueConstraintQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraint(t *pqt.Table) {
	g.g.RepositoryUpdateOneByUniqueConstraint(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpsertQuery(t *pqt.Table) {
	g.g.RepositoryUpsertQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpsert(t *pqt.Table) {
	g.g.RepositoryUpsert(t)
	g.g.NewLine()
}

func (g *Generator) generateWhereClause(t *pqt.Table) {
	g.g.WhereClause(t)
	g.g.NewLine()
}

func (g *Generator) generateJoinClause() {
	g.g.JoinClause()
}

func (g *Generator) generateLogFunc(s *pqt.Schema) {
	g.p.Printf(`
	// %s represents function that can be passed into repository to log query result.
	type LogFunc func(err error, ent, fnc, sql string, args ...interface{})`,
		g.Formatter.Identifier("log", "func"),
	)
}

func (g *Generator) generateInterfaces(s *pqt.Schema) {
	g.p.Print(`
	// Rows ...
	type Rows interface {
		io.Closer
		ColumnTypes() ([]*sql.ColumnType, error)
		Columns() ([]string, error)
		Err() error
		Next() bool
		NextResultSet() bool
		Scan(dest ...interface{}) error
	}`)
}

func (g *Generator) generateRepositoryFindQuery(t *pqt.Table) {
	g.g.RepositoryFindQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFind(t *pqt.Table) {
	g.g.RepositoryFind(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindIter(t *pqt.Table) {
	g.g.RepositoryFindIter(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryCount(t *pqt.Table) {
	g.g.RepositoryCount(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryFindOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindOneByUniqueConstraint(t *pqt.Table) {
	g.g.RepositoryFindOneByUniqueConstraint(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryDeleteOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryDeleteOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateEntityProp(t *pqt.Table) {
	g.g.EntityProp(t)
	g.g.NewLine()
}

func (g *Generator) generateEntityProps(t *pqt.Table) {
	g.g.EntityProps(t)
	g.g.NewLine()
}

func (g *Generator) generateScanRows(t *pqt.Table) {
	g.g.ScanRows(t)
	g.g.NewLine()
}

func (g *Generator) generateStatics(s *pqt.Schema) {
	g.p.Print(`

const (
	JoinDoNot = iota
	JoinInner
	JoinLeft
	JoinRight
	JoinCross
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
}`)

	for _, plugin := range g.Plugins {
		if txt := plugin.Static(s); txt != "" {
			g.p.Print(txt)
			g.p.Print("\n\n")
		}
	}
}