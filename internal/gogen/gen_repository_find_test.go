package gogen_test

import (
	"testing"

	"time"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestGenerator_RepositoryFindQuery(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	t2 := pqt.NewTable("t2").
		AddColumn(pqt.NewColumn("xyz", pqtgo.TypeCustom(time.Now(), time.Now(), time.Now()))).
		AddColumn(pqt.NewColumn("abc", pqt.TypeInteger())).
		AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Repository(t2)
	g.RepositoryFindQuery(t2)
	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) FindQuery(fe *T2FindExpr) (string, []interface{}, error) {
	comp := NewComposer(3)
	buf := bytes.NewBufferString("SELECT ")
	if len(fe.Columns) == 0 {
		buf.WriteString("t0.abc, t0.t1_id, t0.xyz")
	} else {
		buf.WriteString(strings.Join(fe.Columns, ", "))
	}
	if fe.JoinT1 != nil && fe.JoinT1.Fetch {
		buf.WriteString(", t1.age, t1.id")
	}
	buf.WriteString(" FROM ")
	buf.WriteString(r.Table)
	buf.WriteString(" AS t0")
	if fe.JoinT1 != nil {
		joinClause(comp, fe.JoinT1.Kind, "t1 AS t1 ON t0.t1_id=t1.id")
		if fe.JoinT1.On != nil {
			comp.Dirty = true
			if err := T1CriteriaWhereClause(comp, fe.JoinT1.On, 1); err != nil {
				return "", nil, err
			}
		}
	}
	if comp.Dirty {
		buf.ReadFrom(comp)
		comp.Dirty = false
	}
	if fe.Where != nil {
		if err := T2CriteriaWhereClause(comp, fe.Where, 0); err != nil {
			return "", nil, err
		}
	}
	if fe.JoinT1 != nil && fe.JoinT1.Where != nil {
		if err := T1CriteriaWhereClause(comp, fe.JoinT1.Where, 1); err != nil {
			return "", nil, err
		}
	}
	if comp.Dirty {
		if _, err := buf.WriteString(" WHERE "); err != nil {
			return "", nil, err
		}
		buf.ReadFrom(comp)
	}

	if len(fe.OrderBy) > 0 {
		i := 0
		for _, order := range fe.OrderBy {
			for _, columnName := range TableT2Columns {
				if order.Name == columnName {
					if i == 0 {
						comp.WriteString(" ORDER BY ")
					}
					if i > 0 {
						if _, err := comp.WriteString(", "); err != nil {
							return "", nil, err
						}
					}
					if _, err := comp.WriteString(order.Name); err != nil {
						return "", nil, err
					}
					if order.Descending {
						if _, err := comp.WriteString(" DESC"); err != nil {
							return "", nil, err
						}
					}
					i++
					break
				}
			}
		}
	}
	if fe.Offset > 0 {
		if _, err := comp.WriteString(" OFFSET "); err != nil {
			return "", nil, err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		if _, err := comp.WriteString(" "); err != nil {
			return "", nil, err
		}
		comp.Add(fe.Offset)
	}
	if fe.Limit > 0 {
		if _, err := comp.WriteString(" LIMIT "); err != nil {
			return "", nil, err
		}
		if err := comp.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		if _, err := comp.WriteString(" "); err != nil {
			return "", nil, err
		}
		comp.Add(fe.Limit)
	}

	buf.ReadFrom(comp)

	return buf.String(), comp.Args(), nil
}`)

}
