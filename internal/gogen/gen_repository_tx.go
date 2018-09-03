package gogen

import (
	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtfmt"
)

func (g *Generator) RepositoryTx(t *pqt.Table) {
	g.Printf(`
type %sRepositoryBaseTx struct {
	base *%sRepositoryBase
	tx *sql.Tx
}`,
		pqtfmt.Public(t.Name),
		pqtfmt.Public(t.Name),
	)
}

func (g *Generator) RepositoryTxMethodCommitMethod(t *pqt.Table) {
	g.Printf(`
func (r %sRepositoryBaseTx) Commit() error {
	return r.tx.Commit()
}`,
		pqtfmt.Public(t.Name),
	)
}

func (g *Generator) RepositoryTxMethodRollbackMethod(t *pqt.Table) {
	g.Printf(`
func (r %sRepositoryBaseTx) Rollback() error {
	return r.tx.Rollback()
}`,
		pqtfmt.Public(t.Name),
	)
}
