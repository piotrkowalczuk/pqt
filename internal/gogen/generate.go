package gogen

import (
	"github.com/piotrkowalczuk/pqt/internal/print"
)

// Package generates package header.
func Package(p *print.Printer, pkg string) {
	if pkg == "" {
		pkg = "main"
	}
	p.Printf("package %s\n", pkg)
}
