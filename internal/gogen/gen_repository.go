package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) Repository(t *pqt.Table) {
	g.Printf(`
type %sRepositoryBase struct {
	%s string
	%s []string
	%s *sql.DB
	%s LogFunc
}`,
		pqtfmt.Public(t.Name),
		pqtfmt.Public("table"),
		pqtfmt.Public("columns"),
		pqtfmt.Public("db"),
		pqtfmt.Public("log"),
	)
}

func (g *Generator) RepositoryMethodTx(t *pqt.Table) {
	g.Printf(`
		func (r *%sRepositoryBase) %s(tx *sql.Tx) (*%sRepositoryBaseTx, error) {`,
		pqtfmt.Public(t.Name),
		pqtfmt.Public("tx"),
		pqtfmt.Public(t.Name),
	)
	g.Printf(`
	return &%sRepositoryBaseTx{
		base: r,
		tx: tx,
	}, nil
}`,
		pqtfmt.Public(t.Name),
	)
}

func (g *Generator) RepositoryMethodBeginTx(t *pqt.Table) {
	g.Printf(`
		func (r *%sRepositoryBase) %s(ctx context.Context) (*%sRepositoryBaseTx, error) {`,
		pqtfmt.Public(t.Name),
		pqtfmt.Public("beginTx"),
		pqtfmt.Public(t.Name),
	)
	g.Printf(`
	tx, err := r.%s.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return r.%s(tx)
}`,
		pqtfmt.Public("db"),
		pqtfmt.Public("tx"),
	)
}

func (g *Generator) RepositoryMethodRunInTransaction(t *pqt.Table) {
	g.Printf(`
func (r %sRepositoryBase) RunInTransaction(ctx context.Context, fn func(rtx *%sRepositoryBaseTx) error, attempts int) (err error) {
	return RunInTransaction(ctx, r.%s, func(tx *sql.Tx) error {
		rtx, err := r.Tx(tx)
		if err != nil {
			return err
		}
		return fn(rtx)
	}, attempts)
}`,
		pqtfmt.Public(t.Name),
		pqtfmt.Public(t.Name),
		pqtfmt.Public("db"),
	)
}
