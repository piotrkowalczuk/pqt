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
				).AddColumn(
					pqt.NewColumn("tags", pqt.TypeTextArray(0)),
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
		tableFirstColumnTags = "tags"
		)
var (
tableFirstColumns = []string{
tableFirstColumnId,
tableFirstColumnName,
tableFirstColumnTags,
})
type firstEntity struct{
// id ...
id *ntypes.Int64
// name ...
name *ntypes.String
// tags ...
tags pq.StringArray
}

func (e *firstEntity) prop(cn string) (interface{}, bool) {
switch cn {
case tableFirstColumnId:
return &e.id, true
case tableFirstColumnName:
return &e.name, true
case tableFirstColumnTags:
return &e.tags, true
default:
return nil, false
}
}
func (e *firstEntity) props(cns ...string) ([]interface{}, error) {

		res := make([]interface{}, 0, len(cns))
		for _, cn := range cns {
			if prop, ok := e.prop(cn); ok {
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

// Ent is wrapper around first method that makes iterator more generic.
func (i *firstIterator) Ent() (interface{}, error) {
	return i.First()
}

func (i *firstIterator) First() (*firstEntity, error) {
	var ent firstEntity
	cols, err := i.rows.Columns()
	if err != nil {
		return nil, err
	}

	props, err := ent.props(cols...)
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
tags pq.StringArray
}

func (c *firstCriteria) WriteComposition(sel string, com *pqtgo.Composer, opt *pqtgo.CompositionOpts) (err error) {
	
		if err = pqtgo.WriteCompositionQueryInt64(c.id, tableFirstColumnId, com, &pqtgo.CompositionOpts{
		Joint: " AND ",
		IsJSON: false,
	}); err != nil {
			return
		}

		if err = pqtgo.WriteCompositionQueryString(c.name, tableFirstColumnName, com, pqtgo.And); err != nil {
			return
		}
 if c.tags != nil {if com.Dirty {
		com.WriteString(" AND ")
	}
	com.Dirty = true
if _, err = com.WriteString(tableFirstColumnTags); err != nil {
			return
		}
		if _, err = com.WriteString(" = "); err != nil {
			return
		}
		if err = com.WritePlaceholder(); err != nil {
			return
		}
		
if com.Dirty {
			if opt.Cast != "" {
				if _, err = com.WriteString(opt.Cast); err != nil {
					return
				}
			} else {
				if _, err = com.WriteString(" "); err != nil {
					return
				}
			}
		}
		
com.Add(c.tags)
		}

	if len(c.sort) > 0 {
		i:=0
		com.WriteString(" ORDER BY ")

		for cn, asc := range c.sort {
			for _, tcn := range tableFirstColumns {
				if cn == tcn {
					if i > 0 {
						com.WriteString(", ")
					}
					com.WriteString(cn)
					if !asc {
						com.WriteString(" DESC ")
					}
					i++
					break
				}
			}
		}
	}
	if c.offset > 0 {
		if _, err = com.WriteString(" OFFSET "); err != nil {
			return
		}
		if err = com.WritePlaceholder(); err != nil {
			return
		}
		if _, err = com.WriteString(" "); err != nil {
			return
		}
		com.Add(c.offset)
	}
	if c.limit > 0 {
		if _, err = com.WriteString(" LIMIT "); err != nil {
			return
		}
		if err = com.WritePlaceholder(); err != nil {
			return
		}
		if _, err = com.WriteString(" "); err != nil {
			return
		}
		com.Add(c.limit)
	}

	return
}
type firstPatch struct {
id *ntypes.Int64
name *ntypes.String
tags pq.StringArray
}


		type firstRepositoryBase struct {
			table string
			columns []string
			db *sql.DB
			dbg bool
			log log.Logger
		}
	func scanFirstRows(rows *sql.Rows) ([]*firstEntity, error) {
	var (
		entities []*firstEntity
		err error
	)
	for rows.Next() {
		var ent firstEntity
		err = rows.Scan(
	&ent.id,
&ent.name,
&ent.tags,
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

	func (r *firstRepositoryBase) count(c *firstCriteria) (int64, error) {

	com := pqtgo.NewComposer(3)
	buf := bytes.NewBufferString("SELECT COUNT(*) FROM ")
	buf.WriteString(r.table)

	if err := c.WriteComposition("", com, pqtgo.And); err != nil {
		return 0, err
	}
	if com.Dirty {
		buf.WriteString(" WHERE ")
	}
	if com.Len() > 0 {
		buf.ReadFrom(com)
	}

	if r.dbg {
		if err := r.log.Log("msg", buf.String(), "function", "Count"); err != nil {
			return 0, err
		}
	}

	var count int64
	if err := r.db.QueryRow(buf.String(), com.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *firstRepositoryBase) find(c *firstCriteria) ([]*firstEntity, error) {

	com := pqtgo.NewComposer(1)
	buf := bytes.NewBufferString("SELECT ")
	buf.WriteString(strings.Join(r.columns, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(r.table)
	buf.WriteString(" ")

	if err := c.WriteComposition("", com, pqtgo.And); err != nil {
		return nil, err
	}
	if com.Dirty {
		buf.WriteString(" WHERE ")
	}
	if com.Len() > 0 {
		buf.ReadFrom(com)
	}

	if r.dbg {
		if err := r.log.Log("msg", buf.String(), "function", "Find"); err != nil {
			return nil, err
		}
	}

	rows, err := r.db.Query(buf.String(), com.Args()...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanFirstRows(rows)
}
func (r *firstRepositoryBase) findIter(c *firstCriteria) (*firstIterator, error) {

	com := pqtgo.NewComposer(1)
	buf := bytes.NewBufferString("SELECT ")
	buf.WriteString(strings.Join(r.columns, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(r.table)
	buf.WriteString(" ")

	if err := c.WriteComposition("", com, pqtgo.And); err != nil {
		return nil, err
	}
	if com.Dirty {
		buf.WriteString(" WHERE ")
	}
	if com.Len() > 0 {
		buf.ReadFrom(com)
	}

	if r.dbg {
		if err := r.log.Log("msg", buf.String(), "function", "Find"); err != nil {
			return nil, err
		}
	}

	rows, err := r.db.Query(buf.String(), com.Args()...)
	if err != nil {
		return nil, err
	}


	return &firstIterator{rows: rows}, nil
}
func (r *firstRepositoryBase) insert(e *firstEntity) (*firstEntity, error) {
		insert := pqcomp.New(0, 3)
	insert.AddExpr(tableFirstColumnName, "", e.name)
insert.AddExpr(tableFirstColumnTags, "", e.tags)

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
				b.WriteString(" RETURNING ")
				b.WriteString(strings.Join(r.columns, ", "))
			}
		}

		if r.dbg {
			if err := r.log.Log("msg", b.String(), "function", "Insert"); err != nil {
				return nil, err
			}
		}

		err := r.db.QueryRow(b.String(), insert.Args()...).Scan(
	&e.id,
&e.name,
&e.tags,
)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
func (r *firstRepositoryBase) upsert(e *firstEntity, p *firstPatch, inf ...string) (*firstEntity, error) {
		insert := pqcomp.New(0, 3)
		update := insert.Compose(3)
	insert.AddExpr(tableFirstColumnName, "", e.name)
insert.AddExpr(tableFirstColumnTags, "", e.tags)
if len(inf) > 0 {
update.AddExpr(tableFirstColumnName, "=", p.name)
update.AddExpr(tableFirstColumnTags, "=", p.tags)
}

		b := bytes.NewBufferString("INSERT INTO " + r.table)

		if insert.Len() > 0 {
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
		}
		b.WriteString(" ON CONFLICT ")
		if len(inf) > 0 && update.Len() > 0 {
			b.WriteString(" (")
			for j, i := range inf {
				if j != 0 {
					b.WriteString(", ")
				}
				b.WriteString(i)
			}
			b.WriteString(") ")
			b.WriteString(" DO UPDATE SET ")
			for update.Next() {
				if !update.First() {
					b.WriteString(", ")
				}

				b.WriteString(update.Key())
				b.WriteString(" ")
				b.WriteString(update.Oper())
				b.WriteString(" ")
				b.WriteString(update.PlaceHolder())
			}
		} else {
			b.WriteString(" DO NOTHING ")
		}
		if insert.Len() > 0 {
			if len(r.columns) > 0 {
				b.WriteString(" RETURNING ")
				b.WriteString(strings.Join(r.columns, ", "))
			}
		}

		if r.dbg {
			if err := r.log.Log("msg", b.String(), "function", "Upsert"); err != nil {
				return nil, err
			}
		}

		err := r.db.QueryRow(b.String(), insert.Args()...).Scan(
	&e.id,
&e.name,
&e.tags,
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

func assertGoCode(t *testing.T, s1, s2, msg string, com ...interface{}) {
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
