package gogen_test

import (
	"testing"

	"time"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/testutil"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestGenerator_RepositoryFindIter(t *testing.T) {
	t1 := pqt.NewTable("t1")
	g := &gogen.Generator{}
	g.Repository(t1)
	g.RepositoryFindIter(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) FindIter(ctx context.Context, fe *T1FindExpr) (*T1Iterator, error) {
	query, args, err := r.FindQuery(fe)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if r.Log != nil {
		r.Log(err, TableT1, "find iter", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return &T1Iterator{
		rows: rows,
		expr: fe,
		cols: fe.Columns,
	}, nil
}`)
}

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
	testutil.AssertOutput(t, g.Printer, `
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
	if fe.JoinT1 != nil && fe.JoinT1.Kind.Actionable() && fe.JoinT1.Fetch {
		buf.WriteString(", t1.age, t1.id")
	}
	buf.WriteString(" FROM ")
	buf.WriteString(r.Table)
	buf.WriteString(" AS t0")
	if fe.JoinT1 != nil && fe.JoinT1.Kind.Actionable(){
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
	if fe.JoinT1 != nil && fe.JoinT1.Kind.Actionable() && fe.JoinT1.Where != nil {
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

func TestGenerator_RepositoryFindOneByPrimaryKey(t *testing.T) {
	t1 := pqt.NewTable("t1")
	g := &gogen.Generator{}
	g.Repository(t1)
	g.RepositoryFindOneByPrimaryKey(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)

	t1.AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	g.Reset()
	g.Repository(t1)
	g.RepositoryFindOneByPrimaryKey(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) FindOneByID(ctx context.Context, pk int64) (*T1Entity, error) {
	find := NewComposer(2)
	find.WriteString("SELECT ")
	if len(r.Columns) == 0 {
		find.WriteString("age, id")
	} else {
		find.WriteString(strings.Join(r.Columns, ", "))
	}
	find.WriteString(" FROM ")
	find.WriteString(TableT1)
	find.WriteString(" WHERE ")
	find.WriteString(TableT1ColumnID)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(pk)
	var (
		ent T1Entity
	)
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)
	if r.Log != nil {
		r.Log(err, TableT1, "find by primary key", find.String(), find.Args()...)
	}
	if err != nil {
		return nil, err
	}
	return &ent, nil
}`)
}

func TestGenerator_RepositoryFind(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	g := &gogen.Generator{}
	g.Reset()
	g.Repository(t1)
	g.RepositoryFind(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) Find(ctx context.Context, fe *T1FindExpr) ([]*T1Entity, error) {
	query, args, err := r.FindQuery(fe)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if r.Log != nil {
		r.Log(err, TableT1, "find", query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		entities []*T1Entity
		props []interface{}
	)
	for rows.Next() {
		var ent T1Entity
		if props, err = ent.Props(); err != nil {
			return nil, err
		}
		err = rows.Scan(props...)
		if err != nil {
			return nil, err
		}

		entities = append(entities, &ent)
	}
	err = rows.Err()
	if r.Log != nil {
		r.Log(err, TableT1, "find", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return entities, nil
}`)
}
func TestGenerator_RepositoryFindOneByUniqueConstraint(t *testing.T) {
	firstName := pqt.NewColumn("first_name", pqt.TypeText(), pqt.WithNotNull())
	age := pqt.NewColumn("age", pqt.TypeIntegerBig(), pqt.WithNotNull())
	lastName := pqt.NewColumn("last_name", pqt.TypeText(), pqt.WithNotNull())
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(age).
		AddColumn(firstName).
		AddColumn(lastName).
		AddUnique(firstName, lastName).
		AddUniqueIndex("AgeIsGreaterThanZero", "age>0", firstName, lastName, age)

	g := &gogen.Generator{}
	g.Reset()
	g.Repository(t1)
	g.RepositoryFindOneByUniqueConstraint(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) FindOneByFirstNameAndLastName(ctx context.Context, t1FirstName string, t1LastName string) (*T1Entity, error) {
	find := NewComposer(4)
	find.WriteString("SELECT ")
	if len(r.Columns) == 0 {
		find.WriteString("age, first_name, id, last_name")
	} else {
		find.WriteString(strings.Join(r.Columns, ", "))
	}
	find.WriteString(" FROM ")
	find.WriteString(TableT1)
	find.WriteString(" WHERE ")
	find.WriteString(TableT1ColumnFirstName)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(t1FirstName)
	find.WriteString(" AND ")
	find.WriteString(TableT1ColumnLastName)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(t1LastName)

	var (
		ent T1Entity
	)
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)
	if err != nil {
		return nil, err
	}

	return &ent, nil
}

func (r *T1RepositoryBase) FindOneByFirstNameAndLastNameAndAgeWhereAgeIsGreaterThanZero(ctx context.Context, t1FirstName string, t1LastName string, t1Age int64) (*T1Entity, error) {
	find := NewComposer(4)
	find.WriteString("SELECT ")
	if len(r.Columns) == 0 {
		find.WriteString("age, first_name, id, last_name")
	} else {
		find.WriteString(strings.Join(r.Columns, ", "))
	}
	find.WriteString(" FROM ")
	find.WriteString(TableT1)
	find.WriteString(" WHERE age>0 AND ")
	find.WriteString(TableT1ColumnFirstName)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(t1FirstName)
	find.WriteString(" AND ")
	find.WriteString(TableT1ColumnLastName)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(t1LastName)
	find.WriteString(" AND ")
	find.WriteString(TableT1ColumnAge)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(t1Age)

	var (
		ent T1Entity
	)
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)
	if err != nil {
		return nil, err
	}

	return &ent, nil
}`)
}
