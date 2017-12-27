package pqtgogen

import (
	"go/format"
	"io"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
)

type Generator struct {
	Formatter  *Formatter
	Version    float64
	Pkg        string
	Imports    []string
	Plugins    []Plugin
	Components Component

	g *gogen.Generator
	p *print.Printer
}

// Generate ...
func (g *Generator) Generate(s *pqt.Schema) ([]byte, error) {
	if err := g.generate(s); err != nil {
		return nil, err
	}

	return format.Source(g.p.Bytes())
}

// GenerateTo ...
func (g *Generator) GenerateTo(w io.Writer, s *pqt.Schema) error {
	if err := g.generate(s); err != nil {
		return err
	}

	buf, err := format.Source(g.p.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func (g *Generator) generate(s *pqt.Schema) error {
	g.g = &gogen.Generator{
		Version: g.Version,
	}
	for _, p := range g.Plugins {
		g.g.Plugins = append(g.g.Plugins, p)
	}
	g.p = &g.g.Printer

	g.generatePackage()
	g.generateImports(s)
	if g.Components&ComponentRepository != 0 {
		g.generateLogFunc(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 || g.Components&ComponentHelpers != 0 {
		g.generateInterfaces(s)
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
		g.generateJoinClause()
	}
	for _, t := range s.Tables {
		g.generateConstantsAndVariables(t)
		g.generateEntity(t)
		g.generateEntityProp(t)
		g.generateEntityProps(t)
		if g.Components&ComponentHelpers != 0 {
			g.generateScanRows(t)
		}
		if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
			g.generateIterator(t)
			g.generateCriteria(t)
			g.generateFindExpr(t)
			g.generateJoin(t)
		}
		if g.Components&ComponentCount != 0 {
			g.generateCountExpr(t)
		}
		if g.Components&ComponentUpdate != 0 || g.Components&ComponentUpsert != 0 {
			g.generatePatch(t)
		}
		if g.Components&ComponentRepository != 0 {
			g.generateRepository(t)
			if g.Components&ComponentInsert != 0 {
				g.generateRepositoryInsertQuery(t)
				g.generateRepositoryInsert(t)
			}
			if g.Components&ComponentFind != 0 {
				g.generateWhereClause(t)
				g.generateRepositoryFindQuery(t)
				g.generateRepositoryFind(t)
				g.generateRepositoryFindIter(t)
				g.generateRepositoryFindOneByPrimaryKey(t)
				g.generateRepositoryFindOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpdate != 0 {
				g.generateRepositoryUpdateOneByPrimaryKeyQuery(t)
				g.generateRepositoryUpdateOneByPrimaryKey(t)
				g.generateRepositoryUpdateOneByUniqueConstraintQuery(t)
				g.generateRepositoryUpdateOneByUniqueConstraint(t)
			}
			if g.Components&ComponentUpsert != 0 {
				g.generateRepositoryUpsertQuery(t)
				g.generateRepositoryUpsert(t)
			}
			if g.Components&ComponentCount != 0 {
				g.generateRepositoryCount(t)
			}
			if g.Components&ComponentDelete != 0 {
				g.generateRepositoryDeleteOneByPrimaryKey(t)
			}
		}
	}
	g.generateStatics(s)

	return g.p.Err
}

func (g *Generator) generatePackage() {
	g.g.Package(g.Pkg)
}

func (g *Generator) generateImports(s *pqt.Schema) {
	g.g.Imports(s, "github.com/m4rw3r/uuid")
}

func (g *Generator) generateEntity(t *pqt.Table) {
	g.g.Entity(t)
	g.g.NewLine()
}

func (g *Generator) generateFindExpr(t *pqt.Table) {
	g.g.FindExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCountExpr(t *pqt.Table) {
	g.g.CountExpr(t)
	g.g.NewLine()
}

func (g *Generator) generateCriteria(t *pqt.Table) {
	g.g.Criteria(t)
	g.g.NewLine()
	g.g.Operand(t)
	g.g.NewLine()
}

func (g *Generator) generateJoin(t *pqt.Table) {
	g.g.Join(t)
	g.g.NewLine()
}

func (g *Generator) generatePatch(t *pqt.Table) {
	g.g.Patch(t)
	g.g.NewLine()
}

func (g *Generator) generateIterator(t *pqt.Table) {
	g.g.Iterator(t)
	g.g.NewLine()
}

func (g *Generator) generateRepository(t *pqt.Table) {
	g.g.Repository(t)
	g.g.NewLine()
}

func (g *Generator) generateConstantsAndVariables(t *pqt.Table) {
	g.g.Constraints(t)
	g.g.NewLine()
	g.g.Columns(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsertQuery(t *pqt.Table) {
	g.g.RepositoryInsertQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryInsert(t *pqt.Table) {
	g.g.RepositoryInsert(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKeyQuery(t *pqt.Table) {
	g.g.RepositoryUpdateOneByPrimaryKeyQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryUpdateOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraintQuery(t *pqt.Table) {
	g.g.RepositoryUpdateOneByUniqueConstraintQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpdateOneByUniqueConstraint(t *pqt.Table) {
	g.g.RepositoryUpdateOneByUniqueConstraint(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpsertQuery(t *pqt.Table) {
	g.g.RepositoryUpsertQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryUpsert(t *pqt.Table) {
	g.g.RepositoryUpsert(t)
	g.g.NewLine()
}

func (g *Generator) generateWhereClause(t *pqt.Table) {
	g.g.WhereClause(t)
	g.g.NewLine()
}

func (g *Generator) generateJoinClause() {
	g.g.JoinClause()
}

func (g *Generator) generateLogFunc(s *pqt.Schema) {
	g.p.Printf(`
	// %s represents function that can be passed into repository to log query result.
	type LogFunc func(err error, ent, fnc, sql string, args ...interface{})`,
		g.Formatter.Identifier("log", "func"),
	)
}

func (g *Generator) generateInterfaces(s *pqt.Schema) {
	g.g.Interfaces()
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindQuery(t *pqt.Table) {
	g.g.RepositoryFindQuery(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFind(t *pqt.Table) {
	g.g.RepositoryFind(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindIter(t *pqt.Table) {
	g.g.RepositoryFindIter(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryCount(t *pqt.Table) {
	g.g.RepositoryCount(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryFindOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryFindOneByUniqueConstraint(t *pqt.Table) {
	g.g.RepositoryFindOneByUniqueConstraint(t)
	g.g.NewLine()
}

func (g *Generator) generateRepositoryDeleteOneByPrimaryKey(t *pqt.Table) {
	g.g.RepositoryDeleteOneByPrimaryKey(t)
	g.g.NewLine()
}

func (g *Generator) generateEntityProp(t *pqt.Table) {
	g.g.EntityProp(t)
	g.g.NewLine()
}

func (g *Generator) generateEntityProps(t *pqt.Table) {
	g.g.EntityProps(t)
	g.g.NewLine()
}

func (g *Generator) generateScanRows(t *pqt.Table) {
	g.g.ScanRows(t)
	g.g.NewLine()
}

func (g *Generator) generateStatics(s *pqt.Schema) {
	g.g.Statics(s)
	g.g.NewLine()
}
