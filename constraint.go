package pqt

import (
	"fmt"
	"strings"
)

const (
	// ConstraintTypeUnknown ...
	ConstraintTypeUnknown = "unknown"
	// ConstraintTypePrimaryKey ...
	ConstraintTypePrimaryKey = "pkey"
	// ConstraintTypeCheck ...
	ConstraintTypeCheck = "check"
	// ConstraintTypeUnique ...
	ConstraintTypeUnique = "key"
	// ConstraintTypeIndex ...
	ConstraintTypeIndex = "idx"
	// ConstraintTypeForeignKey ...
	ConstraintTypeForeignKey = "fkey"
	// ConstraintTypeExclusion ...
	ConstraintTypeExclusion = "excl"
)

// ConstraintOption ...
type ConstraintOption func(*Constraint)

// Constraint ...
type Constraint struct {
	Type, Check                                                          string
	Table, ReferenceTable                                                *Table
	Columns, ReferenceColumns                                            Columns
	Attribute                                                            []*Attribute
	Match, OnDelete, OnUpdate                                            int32
	NoInherit, DeferrableInitiallyDeferred, DeferrableInitiallyImmediate bool
}

// Name ...
func (c *Constraint) Name() string {
	var schema string

	switch {
	case c.Table == nil:
		return "<missing table>"
	case c.Table.Schema == nil || c.Table.Schema.Name == "":
		schema = "public"
	default:
		schema = c.Table.Schema.Name
	}

	if len(c.Columns) == 0 {
		return fmt.Sprintf("%s.%s_%s", schema, c.Table.ShortName, c.Type)
	}
	tmp := make([]string, 0, len(c.Columns))
	for _, col := range c.Columns {
		if col.ShortName != "" {
			tmp = append(tmp, col.ShortName)
			continue
		}
		tmp = append(tmp, col.Name)
	}

	return fmt.Sprintf("%s.%s_%s_%s", schema, c.Table.ShortName, strings.Join(tmp, "_"), c.Type)
}

// Unique constraint ensure that the data contained in a column or a group of columns is unique with respect to all the rows in the table.
func Unique(table *Table, columns ...*Column) *Constraint {
	return &Constraint{
		Type:    ConstraintTypeUnique,
		Table:   table,
		Columns: columns,
	}
}

// PrimaryKey constraint is simply a combination of a unique constraint and a not-null constraint.
func PrimaryKey(table *Table, columns ...*Column) *Constraint {
	return &Constraint{
		Type:    ConstraintTypePrimaryKey,
		Table:   table,
		Columns: columns,
	}
}

// Check ...
func Check(table *Table, check string, columns ...*Column) *Constraint {
	return &Constraint{
		Type:    ConstraintTypeCheck,
		Table:   table,
		Columns: columns,
		Check:   check,
	}
}

// Exclusion constraint ensure that if any two rows are compared on the specified columns
// or expressions using the specified operators,
// at least one of these operator comparisons will return false or null.
func Exclusion(table *Table, columns ...*Column) *Constraint {
	return &Constraint{
		Type:    ConstraintTypeExclusion,
		Table:   table,
		Columns: columns,
	}
}

// ForeignKey constraint specifies that the values in a column (or a group of columns)
// must match the values appearing in some row of another table.
// We say this maintains the referential integrity between two related tables.
//func ForeignKey(column *Column, opts ...ConstraintOption) *Constraint {
//	fk := &Constraint{
//		Type:    ConstraintTypeForeignKey,
//		Table:   column.Table,
//		Columns: []*Column{column},
//	}
//	for _, o := range opts {
//		o(fk)
//	}
//
//	return fk
//}

// Reference ...
type Reference struct {
	From, To *Column
}

// ForeignKey constraint specifies that the values in a column (or a group of columns)
// must match the values appearing in some row of another table.
// We say this maintains the referential integrity between two related tables.
func ForeignKey(table *Table, columns, references Columns, opts ...ConstraintOption) *Constraint {
	if len(references) == 0 {
		panic("foreign key expects at least one reference column")
	}
	fk := &Constraint{
		Type:             ConstraintTypeForeignKey,
		Columns:          columns,
		Table:            columns[0].Table,
		ReferenceColumns: references,
		ReferenceTable:   references[0].Table,
	}

	for _, o := range opts {
		o(fk)
	}

	return fk
}

// Index ...
func Index(table *Table, columns ...*Column) *Constraint {
	return &Constraint{
		Type:    ConstraintTypeIndex,
		Table:   table,
		Columns: columns,
	}
}

// String implements Stringer interface.
func (c *Constraint) String() string {
	return c.Name()
}

// IsForeignKey returns true if string has suffix "_fkey".
func IsForeignKey(c string) bool {
	return strings.HasSuffix(c, ConstraintTypeForeignKey)
}

// IsUnique returns true if string has suffix "_key".
func IsUnique(c string) bool {
	return strings.HasSuffix(c, ConstraintTypeUnique)
}

// IsPrimaryKey returns true if string has suffix "_pkey".
func IsPrimaryKey(c string) bool {
	return strings.HasSuffix(c, ConstraintTypePrimaryKey)
}

// IsCheck returns true if string has suffix "_check".
func IsCheck(c string) bool {
	return strings.HasSuffix(c, ConstraintTypeCheck)
}

// IsExclusion returns true if string has suffix "_excl".
func IsExclusion(c string) bool {
	return strings.HasSuffix(c, ConstraintTypeExclusion)
}

// IsIndex returns true if string has suffix "_idx".
func IsIndex(c string) bool {
	return strings.HasSuffix(c, ConstraintTypeIndex)
}

//
//// WithReference ...
//func WithReference(columns ...*Column) ConstraintOption {
//	return func(constraint *Constraint) {
//		var referenceTable *Table
//
//		for i, c := range columns {
//			if i == 0 {
//				referenceTable = c.Table
//				continue
//			}
//
//			if referenceTable != c.Table {
//				panic("pqt: each column in foreign key needs to be in the same table")
//			}
//		}
//
//		constraint.ReferenceColumns = columns
//	}
//}
