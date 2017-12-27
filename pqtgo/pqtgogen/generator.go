package pqtgogen

import (
	"go/format"
	"io"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
)

type Component uint64

const (
	ComponentInsert Component = 1 << (64 - 1 - iota)
	ComponentFind
	ComponentUpdate
	ComponentUpsert
	ComponentCount
	ComponentDelete
	ComponentHelpers

	ComponentRepository = ComponentInsert | ComponentFind | ComponentUpdate | ComponentUpsert | ComponentCount | ComponentDelete
	ComponentAll        = ComponentRepository | ComponentHelpers
)

type Generator struct {
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

	g.g.Package(g.Pkg)
	g.g.Imports(s, "github.com/m4rw3r/uuid")
	if g.Components&ComponentRepository != 0 {
		g.g.Funcs()
		g.g.NewLine()
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 || g.Components&ComponentHelpers != 0 {
		g.g.Interfaces()
		g.g.NewLine()
	}
	if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
		g.g.JoinClause()
		g.g.NewLine()
	}
	for _, t := range s.Tables {
		g.g.Constraints(t)
		g.g.NewLine()
		g.g.Columns(t)
		g.g.NewLine()
		g.g.Entity(t)
		g.g.NewLine()
		g.g.EntityProp(t)
		g.g.NewLine()
		g.g.EntityProps(t)
		g.g.NewLine()
		if g.Components&ComponentHelpers != 0 {
			g.g.ScanRows(t)
			g.g.NewLine()
		}
		if g.Components&ComponentFind != 0 || g.Components&ComponentCount != 0 {
			g.g.Iterator(t)
			g.g.NewLine()
			g.g.Criteria(t)
			g.g.NewLine()
			g.g.Operand(t)
			g.g.NewLine()
			g.g.FindExpr(t)
			g.g.NewLine()
			g.g.Join(t)
			g.g.NewLine()
		}
		if g.Components&ComponentCount != 0 {
			g.g.CountExpr(t)
			g.g.NewLine()
		}
		if g.Components&ComponentUpdate != 0 || g.Components&ComponentUpsert != 0 {
			g.g.Patch(t)
			g.g.NewLine()
		}
		if g.Components&ComponentRepository != 0 {
			g.g.Repository(t)
			g.g.NewLine()

			if g.Components&ComponentInsert != 0 {
				g.g.RepositoryInsertQuery(t)
				g.g.NewLine()
				g.g.RepositoryInsert(t)
				g.g.NewLine()
			}
			if g.Components&ComponentFind != 0 {
				g.g.WhereClause(t)
				g.g.NewLine()
				g.g.RepositoryFindQuery(t)
				g.g.NewLine()
				g.g.RepositoryFind(t)
				g.g.NewLine()
				g.g.RepositoryFindIter(t)
				g.g.NewLine()
				g.g.RepositoryFindOneByPrimaryKey(t)
				g.g.NewLine()
				g.g.RepositoryFindOneByUniqueConstraint(t)
				g.g.NewLine()
			}
			if g.Components&ComponentUpdate != 0 {
				g.g.RepositoryUpdateOneByPrimaryKeyQuery(t)
				g.g.NewLine()
				g.g.RepositoryUpdateOneByPrimaryKey(t)
				g.g.NewLine()
				g.g.RepositoryUpdateOneByUniqueConstraintQuery(t)
				g.g.NewLine()
				g.g.RepositoryUpdateOneByUniqueConstraint(t)
				g.g.NewLine()
			}
			if g.Components&ComponentUpsert != 0 {
				g.g.RepositoryUpsertQuery(t)
				g.g.NewLine()
				g.g.RepositoryUpsert(t)
				g.g.NewLine()
			}
			if g.Components&ComponentCount != 0 {
				g.g.RepositoryCount(t)
				g.g.NewLine()
			}
			if g.Components&ComponentDelete != 0 {
				g.g.RepositoryDeleteOneByPrimaryKey(t)
				g.g.NewLine()
			}
		}
	}
	g.g.Statics(s)
	g.g.NewLine()

	return g.p.Err
}
