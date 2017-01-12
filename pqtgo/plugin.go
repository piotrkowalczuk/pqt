package pqtgo

import "github.com/piotrkowalczuk/pqt"

type Plugin interface {
	PropertyType(*pqt.Column, int32) string
	WhereClause(*pqt.Column) string
	// SetClause allow to generate alternative code for column for update queries.
	// Available placeholders:
	//
	// 	{{ .selector }} - property of patch object
	// 	{{ .column }} - const that represents given column
	// 	{{ .composer }} - Composer instance
	SetClause(*pqt.Column) string
	ScanClause(*pqt.Column) string
	Static(*pqt.Schema) string
}
