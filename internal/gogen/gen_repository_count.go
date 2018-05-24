package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) RepositoryCount(t *pqt.Table) {
	entityName := pqtfmt.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCountExpr) (int64, error) {`, entityName, pqtfmt.Public("count"), entityName)
	g.Printf(`
		query, args, err := r.%sQuery(&%sFindExpr{
			%s: c.%s,
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
		%s: c.%s,`, pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)), pqtfmt.Public("join", or(r.InversedName, r.InversedTable.Name)))
	}
	g.Printf(`
		})
		if err != nil {
			return 0, err
		}
		var count int64
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(&count)`,
		pqtfmt.Public("db"),
	)

	g.Printf(`
		if r.%s != nil {
			r.%s(err, Table%s, "count", query, args...)
		}
		if err != nil {
			return 0, err
		}
		return count, nil
	}`,
		pqtfmt.Public("log"),
		pqtfmt.Public("log"),
		entityName,
	)
}
