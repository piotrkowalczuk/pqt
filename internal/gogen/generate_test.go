package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt/internal/gogen"
	"github.com/piotrkowalczuk/pqt/internal/print"
)

func TestPackage(t *testing.T) {
	cases := []struct {
		name, exp string
	}{
		{
			name: "something",
			exp:  "package something\n",
		},
		{
			name: "",
			exp:  "package main\n",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			var p print.Printer

			gogen.Package(&p, c.name)

			if p.String() != c.exp {
				t.Errorf("wrong output, expected '%s' but got '%s'", c.exp, p.String())
			}
		})
	}
}
