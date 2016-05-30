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

type firstCriteria struct {
offset, limit int64
sort map[string]bool
id *qtypes.Int64
name *qtypes.String
}


		type firstRepositoryBase struct {
			table string
			columns []string
			db *sql.DB
			dbg bool
			log log.Logger
		}
	func (r *firstRepositoryBase) Find(c *firstCriteria) ([]*firstEntity, error) {
	wbuf := bytes.NewBuffer(nil)
			qbuf := bytes.NewBuffer(nil)
			qbuf.WriteString("SELECT ")
			qbuf.WriteString(strings.Join(r.columns, ", "))
			qbuf.WriteString(" FROM ")
			qbuf.WriteString(r.table)

			pw := pqtgo.NewPlaceholderWriter()
			args := pqtgo.NewArguments(0)
			dirty := false
	
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
						wbuf.WriteString("=")
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_NOT_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" <> ")
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_GREATER:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" > ")
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_GREATER_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" >= ")
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_LESS:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" < ")
						pw.WriteTo(wbuf)
						args.Add(c.id)
					case qtypes.NumericQueryType_LESS_EQUAL:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" >= ")
						pw.WriteTo(wbuf)
						args.Add(c.id.Value())
					case qtypes.NumericQueryType_IN:
						if len(c.id.Values) >0 {
							if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
							wbuf.WriteString(tableFirstColumnId)
							wbuf.WriteString(" IN (")
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
						wbuf.WriteString(" > ")
						pw.WriteTo(wbuf)
						args.Add(c.id.Values[0])
						wbuf.WriteString(" AND ")
						wbuf.WriteString(tableFirstColumnId)
						wbuf.WriteString(" < ")
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
						wbuf.WriteString("=")
						pw.WriteTo(wbuf)
						args.Add(c.name.Value())
					case qtypes.TextQueryType_SUBSTRING:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnName)
						wbuf.WriteString(" LIKE ")
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%%%s%%", c.name.Value()))
					case qtypes.TextQueryType_HAS_PREFIX:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnName)
						wbuf.WriteString(" LIKE ")
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%s%%", c.name.Value()))
					case qtypes.TextQueryType_HAS_SUFFIX:
						if dirty {
		wbuf.WriteString(" AND ")
	}
	dirty = true
	
						wbuf.WriteString(tableFirstColumnName)
						wbuf.WriteString(" LIKE ")
						pw.WriteTo(wbuf)
						args.Add(fmt.Sprintf("%%%s", c.name.Value()))
					}
				}


	if dirty {
		if _, err := qbuf.WriteString(" WHERE "); err != nil {
			return nil, err
		}
		if _, err := wbuf.WriteTo(qbuf); err != nil {
			return nil, err
		}
	}

	if c.offset > 0 {
		qbuf.WriteString(" OFFSET ")
		pw.WriteTo(qbuf)
		args.Add(c.offset)
	}
	if c.limit > 0 {
		qbuf.WriteString(" LIMIT ")
		pw.WriteTo(qbuf)
		args.Add(c.limit)
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

	var entities []*firstEntity
	for rows.Next() {
		var entity firstEntity
		err = rows.Scan(
	&entity.Id,
&entity.Name,
)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &entity)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		return entities, nil
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
