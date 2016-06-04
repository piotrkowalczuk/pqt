package pqtgo_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestGenerator_Generate(t *testing.T) {
	cases := map[string]struct {
		schema    *pqt.Schema
		generator *pqtgo.Generator
		expected  string
	}{
		"basic": {
			schema: pqt.NewSchema("text"),
			generator: pqtgo.NewGenerator().
				AddImport("github.com/piotrkowalczuk/ntypes"),
			expected: `package main
import (
"github.com/go-kit/kit/log"
"github.com/m4rw3r/uuid"
"github.com/piotrkowalczuk/ntypes"
)
`,
		},
		"custom-package": {
			schema: pqt.NewSchema("text"),
			generator: pqtgo.NewGenerator().
				SetPackage("example").
				AddImport("fmt"),
			expected: `package example
import (
"github.com/go-kit/kit/log"
"github.com/m4rw3r/uuid"
"fmt"
)
`,
		},
		"simple table": {
			schema: pqt.NewSchema("text").AddTable(
				pqt.NewTable("first").AddColumn(
					pqt.NewColumn("id", pqt.TypeSerialBig()),
				).AddColumn(
					pqt.NewColumn("name", pqt.TypeText()),
				),
			),
			generator: pqtgo.NewGenerator().
				SetPackage("custom"),
			expected: `package custom
import (
"github.com/go-kit/kit/log"
"github.com/m4rw3r/uuid"
)
const (
tableFirst = "text.first"
tableFirstColumnId = "id"
tableFirstColumnName = "name"
)
var (
tableFirstColumns = []string{
tableFirstColumnId,
tableFirstColumnName,
})
type firstEntity struct{
Id *ntypes.Int64
Name *ntypes.String
}

func (e *firstEntity) Prop(cn string) (interface{}, bool) {
switch cn {
case tableFirstColumnId:
return &e.Id, true
case tableFirstColumnName:
return &e.Name, true
default:
return nil, false
}
}
func (e *firstEntity) Props(cns ...string) ([]interface{}, error) {

		res := make([]interface{}, 0, len(cns))
		for _, cn := range cns {
			if prop, ok := e.Prop(cn); ok {
				res = append(res, prop)
			} else {
				return nil, fmt.Errorf("unexpected column provided: %s", cn)
			}
		}
		return res, nil
}


// firstIterator is not thread safe.
type firstIterator struct {
	rows *sql.Rows
	cols []string
}

func (i *firstIterator) Next() bool {
	return i.rows.Next()
}

func (i *firstIterator) Close() error {
	return i.rows.Close()
}

func (i *firstIterator) Err() error {
	return i.rows.Err()
}

// Columns is wrapper around sql.Rows.Columns method, that also cache outpu inside iterator.
func (i *firstIterator) Columns() ([]string, error) {
	if i.cols == nil {
		cols, err := i.rows.Columns()
		if err != nil {
			return nil, err
		}
		i.cols = cols
	}
	return i.cols, nil
}

// Ent is wrapper arround first method that makes iterator more generic.
func (i *firstIterator) Ent() (interface{}, error) {
	return i.First()
}

func (i *firstIterator) First() (*firstEntity, error) {
	var ent firstEntity
	cols, err := i.rows.Columns()
	if err != nil {
		return nil, err
	}

	props, err := ent.Props(cols...)
	if err != nil {
		return nil, err
	}
	if err := i.rows.Scan(props...); err != nil {
		return nil, err
	}
	return &ent, nil
}
type firstCriteria struct {
offset, limit int64
sort map[string]bool
id *qtypes.Int64
name *qtypes.String
}

func (c *firstCriteria) WriteSQL(b *bytes.Buffer, pw *pqtgo.PlaceholderWriter, args *pqtgo.Arguments) (wr int64, err error) {
		var (
			wrt int
			wrt64 int64
			dirty bool
		)

		wbuf := bytes.NewBuffer(nil)

				if c.id != nil && c.id.Valid {
					switch c.id.Type {
					case qtypes.NumericQueryType_NOT_A_NUMBER:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" IS NOT NULL ")
						} else {
							wbuf.WriteString(" IS NULL ")
						}
					case qtypes.NumericQueryType_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" <> ")
						} else {
							wbuf.WriteString("=")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_GREATER:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" <= ")
						} else {
							wbuf.WriteString(" > ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_GREATER_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" < ")
						} else {
							wbuf.WriteString(" >= ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_LESS:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" >= ")
						} else {
							wbuf.WriteString(" < ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id)
					case qtypes.NumericQueryType_LESS_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" > ")
						} else {
							wbuf.WriteString(" <= ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_IN:
						if len(c.id.Values) >0 {
							if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

							wbuf.WriteString(tableFirstColumnId)
							if c.id.Negation {
								wbuf.WriteString(" NOT IN (")
							} else {
								wbuf.WriteString(" IN (")
							}
							for i, v := range c.id.Values {
								if i != 0 {
									wbuf.WriteString(",")
								}
								pw.WriteTo(wbuf)
								args.Add(v)
							}
							wbuf.WriteString(") ")
						}
					case qtypes.NumericQueryType_BETWEEN:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" <= ")
						} else {
							wbuf.WriteString(" > ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Values[0])
						wbuf.WriteString(" AND ")
						wbuf.WriteString(tableFirstColumnId)
						if c.id.Negation {
							wbuf.WriteString(" >= ")
						} else {
							wbuf.WriteString(" < ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.id.Values[1])
					}
				}


				if c.name != nil && c.name.Valid {
					switch c.name.Type {
					case qtypes.TextQueryType_NOT_A_TEXT:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnName)
						if c.name.Negation {
							wbuf.WriteString(" IS NOT NULL ")
						} else {
							wbuf.WriteString(" IS NULL ")
						}
					case qtypes.TextQueryType_EXACT:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnName)
						if c.name.Negation {
							wbuf.WriteString(" <> ")
						} else {
							wbuf.WriteString(" = ")
						}
						pw.WriteTo(wbuf)
						args.Add(c.name.Value())
					case qtypes.TextQueryType_SUBSTRING:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnName)
						if c.name.Negation {
							wbuf.WriteString(" NOT LIKE ")
						} else {
							wbuf.WriteString(" LIKE ")
						}
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%%%s%%", c.name.Value()))
					case qtypes.TextQueryType_HAS_PREFIX:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnName)
						if c.name.Negation {
							wbuf.WriteString(" NOT LIKE ")
						} else {
							wbuf.WriteString(" LIKE ")
						}
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%s%%", c.name.Value()))
					case qtypes.TextQueryType_HAS_SUFFIX:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true

						wbuf.WriteString(tableFirstColumnName)
						if c.name.Negation {
							wbuf.WriteString(" NOT LIKE ")
						} else {
							wbuf.WriteString(" LIKE ")
						}
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%%%s", c.name.Value()))
					}
				}


	if dirty {
		if wrt, err = b.WriteString(" WHERE "); err != nil {
			return
		}
		wr += int64(wrt)
		if wrt64, err = wbuf.WriteTo(b); err != nil {
			return
		}
		wr += wrt64
	}

	if c.offset > 0 {
		b.WriteString(" OFFSET ")
		if wrt64, err = pw.WriteTo(b); err != nil {
			return
		}
		wr += wrt64
		args.Add(c.offset)
	}
	if c.limit > 0 {
		b.WriteString(" LIMIT ")
		if wrt64, err = pw.WriteTo(b); err != nil {
			return
		}
		wr += wrt64
		args.Add(c.limit)
	}

	return
}

		type firstRepositoryBase struct {
			table string
			columns []string
			db *sql.DB
			dbg bool
			log log.Logger
		}
	func ScanFirstRows(rows *sql.Rows) ([]*firstEntity, error) {
	var (
		entities []*firstEntity
		err error
	)
	for rows.Next() {
		var ent firstEntity
		err = rows.Scan(
	&ent.Id,
&ent.Name,
)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &ent)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		return entities, nil
	}

	func (r *firstRepositoryBase) Count(c *firstCriteria) (int64, error) {

	qbuf := bytes.NewBuffer(nil)
	qbuf.WriteString("SELECT COUNT(*) FROM ")
	qbuf.WriteString(r.table)
	pw := pqtgo.NewPlaceholderWriter()
	args := pqtgo.NewArguments(0)

	if _, err := c.WriteSQL(qbuf, pw, args); err != nil {
		return 0, err
	}
	if r.dbg {
		if err := r.log.Log("msg", qbuf.String(), "function", "Count"); err != nil {
			return 0, err
		}
	}

	var count int64
	err := r.db.QueryRow(qbuf.String(), args.Slice()...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (r *firstRepositoryBase) Find(c *firstCriteria) ([]*firstEntity, error) {

	qbuf := bytes.NewBuffer(nil)
	qbuf.WriteString("SELECT ")
	qbuf.WriteString(strings.Join(r.columns, ", "))
	qbuf.WriteString(" FROM ")
	qbuf.WriteString(r.table)

	pw := pqtgo.NewPlaceholderWriter()
	args := pqtgo.NewArguments(0)

	if _, err := c.WriteSQL(qbuf, pw, args); err != nil {
		return nil, err
	}

	if r.dbg {
		if err := r.log.Log("msg", qbuf.String(), "function", "Find"); err != nil {
			return nil, err
		}
	}

	rows, err := r.db.Query(qbuf.String(), args.Slice()...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return ScanFirstRows(rows)
}
func (r *firstRepositoryBase) FindIter(c *firstCriteria) (*firstIterator, error) {

	qbuf := bytes.NewBuffer(nil)
	qbuf.WriteString("SELECT ")
	qbuf.WriteString(strings.Join(r.columns, ", "))
	qbuf.WriteString(" FROM ")
	qbuf.WriteString(r.table)

	pw := pqtgo.NewPlaceholderWriter()
	args := pqtgo.NewArguments(0)

	if _, err := c.WriteSQL(qbuf, pw, args); err != nil {
		return nil, err
	}

	if r.dbg {
		if err := r.log.Log("msg", qbuf.String(), "function", "Find"); err != nil {
			return nil, err
		}
	}

	rows, err := r.db.Query(qbuf.String(), args.Slice()...)
	if err != nil {
		return nil, err
	}


	return &firstIterator{rows: rows}, nil
}
func (r *firstRepositoryBase) Insert(e *firstEntity) (*firstEntity, error) {
		insert := pqcomp.New(0, 2)
	insert.AddExpr(tableFirstColumnName, "", e.Name)

		b := bytes.NewBufferString("INSERT INTO " + r.table)

		if insert.Len() != 0 {
			b.WriteString(" (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.Key())
			}
			insert.Reset()
			b.WriteString(") VALUES (")
			for insert.Next() {
				if !insert.First() {
					b.WriteString(", ")
				}

				fmt.Fprintf(b, "%s", insert.PlaceHolder())
			}
			b.WriteString(")")
			if len(r.columns) > 0 {
				b.WriteString("RETURNING ")
				b.WriteString(strings.Join(r.columns, ","))
			}
		}

		err := r.db.QueryRow(b.String(), insert.Args()...).Scan(
	&e.Id,
&e.Name,
)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
`,
		},
	}

	for hint, c := range cases {
		b, err := c.generator.Generate(c.schema)
		if err != nil {
			t.Errorf("%s: unexpected error: %s", hint, err.Error())
			continue
		}
		assertGoCode(t, c.expected, string(b), hint)
	}
}

func assertGoCode(t *testing.T, s1, s2, msg string, args ...interface{}) {
	s1 = fmt.Sprintf("%s", s1)
	s2 = fmt.Sprintf("%s", s2)
	tmp1 := strings.Split(s1, "\n")
	tmp2 := strings.Split(s2, "\n")
	if s1 != s2 {
		b := bytes.NewBuffer(nil)
		for _, diff := range difflib.Diff(tmp1, tmp2) {
			p := strings.Replace(diff.Payload, "\t", "\\t", -1)
			switch diff.Delta {
			case difflib.Common:
				fmt.Fprintf(b, "%s %s\n", diff.Delta.String(), p)
			case difflib.LeftOnly:
				fmt.Fprintf(b, "\033[31m%s %s\033[39m\n", diff.Delta.String(), p)
			case difflib.RightOnly:
				fmt.Fprintf(b, "\033[32m%s %s\033[39m\n", diff.Delta.String(), p)
			}
		}
		t.Errorf(b.String())
	}
}
