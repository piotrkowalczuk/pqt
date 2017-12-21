package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_RepositoryUpdateOneByPrimaryKey(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").
		AddRelationship(pqt.ManyToOne(t1)).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey()))

	g := &gogen.Generator{}
	g.Repository(t2)
	g.RepositoryUpdateOneByPrimaryKey(t2)
	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) UpdateOneByID(ctx context.Context, pk int64, p *T2Patch) (*T2Entity, error) {
	query, args, err := r.UpdateOneByIDQuery(pk, p)
	if err != nil {
		return nil, err
	}
	var ent T2Entity
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
	if r.Log != nil {
		r.Log(err, TableT2, "update by primary key", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return &ent, nil
}`)
}

func TestGenerator_RepositoryUpdateOneByPrimaryKeyQuery(t *testing.T) {
	t0 := pqt.NewTable("t0")

	g := &gogen.Generator{}
	g.Repository(t0) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByPrimaryKeyQuery(t0)
	assertOutput(t, g.Printer, `
type T0RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)

	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey()))

	pqt.NewSchema("constraints_test").AddTable(t1)

	g = &gogen.Generator{}
	g.Repository(t1) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByPrimaryKeyQuery(t1)

	assertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) UpdateOneByIDQuery(pk int64, p *T1Patch) (string, []interface{}, error) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(r.Table)
	update := NewComposer(1)
	if !update.Dirty {
		return "", nil, errors.New("T1 update failure, nothing to update")
	}
	buf.WriteString(" SET ")
	buf.ReadFrom(update)
	buf.WriteString(" WHERE ")

	update.WriteString(TableT1ColumnID)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(pk)

	buf.ReadFrom(update)
	buf.WriteString(" RETURNING ")
	if len(r.Columns) > 0 {
		buf.WriteString(strings.Join(r.Columns, ", "))
	} else {
		buf.WriteString("id")
	}
	return buf.String(), update.Args(), nil
}`)
}
