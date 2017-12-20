package pqtgogen

import (
	"strings"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

const (
	// Public ...
	Public Visibility = "public"
	// Private ...
	Private Visibility = "private"
)

// Visibility ...
type Visibility string

type Formatter struct {
	Visibility Visibility
	Acronyms   map[string]string
}

func (f *Formatter) Identifier(args ...string) (r string) {
	var vis Visibility
	if f != nil {
		vis = f.Visibility
	}
	var tmp []string
	for _, arg := range args {
		tmp = append(tmp, strings.Split(arg, "_")...)
	}

	switch len(tmp) {
	case 0:
	case 1:
		r = f.identifier(tmp[0], vis)
	default:
		r = f.identifier(tmp[0], vis)
		for _, s := range tmp[1:] {
			r += f.identifier(s, Public)
		}
	}
	return r
}

func (f *Formatter) IdentifierPrivate(args ...string) (r string) {
	switch len(args) {
	case 0:
	case 1:
		r = f.identifier(args[0], Private)
	default:
		r = f.identifier(args[0], Private)
		for _, s := range args[1:] {
			r += f.identifier(s, Public)
		}
	}
	return r
}

func (f *Formatter) identifier(s string, v Visibility) string {
	var acr map[string]string
	if f != nil {
		acr = f.Acronyms
	}
	r := snake(s, v == Private, acr)
	if a, ok := keywords[r]; ok {
		return a
	}
	return r
}

func (f *Formatter) Type(t pqt.Type, m int32) string {
	switch tt := t.(type) {
	case pqt.MappableType:
		for _, mt := range tt.Mapping {
			return f.Type(mt, m)
		}
		return ""
	case pqtgo.BuiltinType:
		return generateTypeBuiltin(tt, m)
	case pqt.BaseType:
		return generateTypeBase(tt, m)
	case pqtgo.CustomType:
		return generateCustomType(tt, m)
	default:
		return ""
	}
}
