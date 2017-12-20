package gogen

import (
	"reflect"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
)

type structField struct {
	Name     string
	Type     string
	Tags     reflect.StructTag
	ReadOnly bool
}

func columnMode(c *pqt.Column, m int32) int32 {
	switch m {
	case pqtgo.ModeCriteria:
	case pqtgo.ModeMandatory:
	case pqtgo.ModeOptional:
	default:
		if c.NotNull || c.PrimaryKey {
			m = pqtgo.ModeMandatory
		}
	}
	return m
}

func or(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
}
