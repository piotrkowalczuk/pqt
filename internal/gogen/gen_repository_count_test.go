package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/testutil"
)

func TestGenerator_RepositoryMethodPrivateCount(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	g := &gogen.Generator{}
	g.Reset()
	g.Repository(t1)
	g.RepositoryMethodPrivateCount(t1)
	testutil.AssertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) count(ctx context.Context, tx *sql.Tx, exp *T1CountExpr) (int64, error) {
	query, args, err := r.FindQuery(&T1FindExpr{
		Where:   exp.Where,
		Columns: []string{"COUNT(*)"},
	})
	if err != nil {
		return 0, err
	}
	var count int64
	if tx == nil {
		err = r.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	}
	if r.Log != nil {
		if tx == nil {
			r.Log(err, TableT1, "count", query, args...)
		} else {
			r.Log(err, TableT1, "count tx", query, args...)
		}
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}`)
}
