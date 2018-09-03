package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func (g *Generator) RepositoryMethodDeleteOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, pk %s) (int64, error) {
			return r.%s(ctx, nil, pk)
		}`,
		entityName,
		pqtfmt.Public("deleteOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		pqtfmt.Private("deleteOneBy", pk.Name),
	)
}

func (g *Generator) RepositoryTxMethodDeleteOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBaseTx) %s(ctx context.Context, pk %s) (int64, error) {
			return r.base.%s(ctx, r.tx, pk)
		}`,
		entityName,
		pqtfmt.Public("deleteOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
		pqtfmt.Private("deleteOneBy", pk.Name),
	)
}

func (g *Generator) RepositoryMethodPrivateDeleteOneByPrimaryKey(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)
	pk, ok := t.PrimaryKey()
	if !ok {
		return
	}

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, pk %s) (int64, error) {`,
		entityName,
		pqtfmt.Private("DeleteOneBy", pk.Name),
		g.columnType(pk, pqtgo.ModeMandatory),
	)
	g.Printf(`
		find := NewComposer(%d)
		find.WriteString("DELETE FROM ")
		find.WriteString(%s)
		find.WriteString(" WHERE ")
		find.WriteString(%s)
		find.WriteString("=")
		find.WritePlaceholder()
		find.Add(pk)`, len(t.Columns),
		pqtfmt.Public("table", t.Name),
		pqtfmt.Public("table", t.Name, "column", pk.Name),
	)

	g.Printf(`
		var (
			err error
			res sql.Result
		)
		if tx == nil {
			res, err = r.%s.ExecContext(ctx, find.String(), find.Args()...)
		} else {
			res, err = tx.ExecContext(ctx, find.String(), find.Args()...)
		}`,
		pqtfmt.Public("db"),
	)
	g.Print(`
		if err != nil {
				return 0, err
			}

		return res.RowsAffected()
	}`)
}
