package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/testutil"
)

func TestGenerator_RepositoryMethodPrivateDeleteOneByPrimaryKey(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	g := &gogen.Generator{}
	g.Reset()
	g.Repository(t1)
	g.RepositoryMethodPrivateDeleteOneByPrimaryKey(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) deleteOneByID(ctx context.Context, tx *sql.Tx, pk int64) (int64, error) {
	find := NewComposer(2)
	find.WriteString("DELETE FROM ")
	find.WriteString(TableT1)
	find.WriteString(" WHERE ")
	find.WriteString(TableT1ColumnID)
	find.WriteString("=")
	find.WritePlaceholder()
	find.Add(pk)
	var (
		err error
		res sql.Result
	)
	if tx == nil {
		res, err = r.DB.ExecContext(ctx, find.String(), find.Args()...)
	} else {
		res, err = tx.ExecContext(ctx, find.String(), find.Args()...)
	}
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}`)
}
