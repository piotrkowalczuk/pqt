package pqtgo

import "github.com/piotrkowalczuk/pqt"

type Plugin interface {
	PropertyType(*pqt.Column, int32) (string)
	WhereClause(*pqt.Column) (string)
}
