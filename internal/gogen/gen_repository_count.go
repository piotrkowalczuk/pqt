package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/formatter"
)

func (g *Generator) RepositoryCount(t *pqt.Table) {
	entityName := formatter.Public(t.Name)

	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context, c *%sCountExpr) (int64, error) {`, entityName, formatter.Public("count"), entityName)
	g.Printf(`
		query, args, err := r.%sQuery(&%sFindExpr{
			%s: c.%s,
			%s: []string{"COUNT(*)"},
		`,
		formatter.Public("find"),
		formatter.Public(entityName),
		formatter.Public("where"),
		formatter.Public("where"),
		formatter.Public("columns"),
	)
	for _, r := range joinableRelationships(t) {
		g.Printf(`
		%s: c.%s,`, formatter.Public("join", or(r.InversedName, r.InversedTable.Name)), formatter.Public("join", or(r.InversedName, r.InversedTable.Name)))
	}
	g.Printf(`
		})
		if err != nil {
			return 0, err
		}
		var count int64
		err = r.%s.QueryRowContext(ctx, query, args...).Scan(&count)`,
		formatter.Public("db"),
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
		formatter.Public("log"),
		formatter.Public("log"),
		entityName,
	)
}
