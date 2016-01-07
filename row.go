package pqt

import "strings"

type Column struct {
	Name, Collate, Default, Check string
	NotNull, Unique, PrimaryKey   bool
	Type                          Type
	Table                         *Table
	Reference                     *Column
}

func (c *Column) Constraint() (*Constraint, bool) {
	var kind string
	switch {
	case c.Unique && !c.PrimaryKey:
		kind = ConstraintTypeUnique
	case c.PrimaryKey && !c.Unique:
		kind = ConstraintTypePrimaryKey
	case c.Check != "":
		kind = ConstraintTypeCheck
	case c.Reference != nil:
		kind = ConstraintTypeForeignKey
	default:
		return nil, false
	}

	cnstr := &Constraint{
		Type:    kind,
		Check:   c.Check,
		Columns: []*Column{c},
		Table:   c.Table,
	}
	if kind == ConstraintTypeForeignKey && c.Reference != nil {
		cnstr.ReferenceColumns = []*Column{c.Reference}
		cnstr.ReferenceTable = c.Reference.Table
	}

	return cnstr, true
}

func JoinColumns(columns []*Column, sep string) string {
	tmp := make([]string, 0, len(columns))
	for _, c := range columns {
		tmp = append(tmp, c.Name)
	}

	return strings.Join(tmp, sep)
}

type Attribute struct {
	Name, Collate, Default, Check string
	NotNull, Unique, PrimaryKey   bool
	Type                          Type
}

func (a *Attribute) Constraint() (*Constraint, bool) {
	var kind string
	switch {
	case a.Unique && !a.PrimaryKey:
		kind = ConstraintTypeUnique
	case a.PrimaryKey && !a.Unique:
		kind = ConstraintTypePrimaryKey
	case a.Check != "":
		kind = ConstraintTypeCheck
	default:
		return nil, false
	}

	return &Constraint{
		Type:      kind,
		Check:     a.Check,
		Attribute: []*Attribute{a},
	}, true
}
