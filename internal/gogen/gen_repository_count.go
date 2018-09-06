package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) RepositoryMethodCount(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, exp *%sCountExpr) (int64, error) {
			return r.%s(ctx, nil, exp)
		}`, entityName, pqtfmt.Public("count"), entityName, pqtfmt.Private("count"))
}

func (g *Generator) RepositoryTxMethodCount(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBaseTx) %s(ctx context.Context, exp *%sCountExpr) (int64, error) {
			return r.base.%s(ctx, r.tx, exp)
		}`, entityName, pqtfmt.Public("count"), entityName, pqtfmt.Private("count"))
}

func (g *Generator) RepositoryMethodPrivateCount(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, tx *sql.Tx, exp *%sCountExpr) (int64, error) {`, entityName, pqtfmt.Private("count"), entityName)
	g.Printf(`
		query, args, err := r.%sQuery(&%sFindExpr{
			%s: exp.%s,
			%s: []string{"COUNT(*)"},
		`,
		pqtfmt.Public("find"),
		pqtfmt.Public(entityName),
		pqtfmt.Public("where"),
		pqtfmt.Public("where"),
		pqtfmt.Public("columns"),
	)
	for _, r := range joinableRelationships(t) {
		g.Printf(`
		%s: exp.%s,`, pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)), pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)))
	}
	g.Printf(`
		})
		if err != nil {
			return 0, err
		}
		var count int64
		if tx == nil {
			err = r.%s.QueryRowContext(ctx, query, args...).Scan(&count)
		} else {
			err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
		}`,
		pqtfmt.Public("db"),
	)

	g.Printf(`
		if r.%s != nil {
			if tx == nil {
				r.%s(err, Table%s, "count", query, args...)
			} else {
				r.%s(err, Table%s, "count tx", query, args...)
			}
		}
		if err != nil {
			return 0, err
		}
		return count, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
		pqtfmt.Public("log"),
		entityName,
	)
}
