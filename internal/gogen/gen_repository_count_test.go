package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_RepositoryCount(t *testing.T) {
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger()))

	g := &gogen.Generator{}
	g.Reset()
	g.Repository(t1)
	g.RepositoryCount(t1)
	assertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) Count(ctx context.Context, c *T1CountExpr) (int64, error) {
	query, args, err := r.FindQuery(&T1FindExpr{
		Where:   c.Where,
		Columns: []string{"COUNT(*)"},
	})
	if err != nil {
		return 0, err
	}
	var count int64
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	if r.Log != nil {
		r.Log(err, TableT1, "count", query, args...)
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}`)
}
