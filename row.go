package pqt

import "strings"

type Column struct {
	Name, Collate, Default, Check string
	NotNull, Unique, PrimaryKey   bool
	Type                          Type
	Table                         *Table
	Reference                     *Column
}

func (c *Column) Constraints() []*Constraint {
	cs := make([]*Constraint, 0)

	if c.Unique && !c.PrimaryKey {
		cs = append(cs, &Constraint{
			Type:    ConstraintTypeUnique,
			Columns: []*Column{c},
			Table:   c.Table,
		})
	}
	if c.PrimaryKey && !c.Unique {
		cs = append(cs, &Constraint{
			Type:    ConstraintTypePrimaryKey,
			Columns: []*Column{c},
			Table:   c.Table,
		})
	}
	if c.Check != "" {
		cs = append(cs, &Constraint{
			Type:    ConstraintTypeCheck,
			Check:   c.Check,
			Columns: []*Column{c},
			Table:   c.Table,
		})
	}
	if c.Reference != nil {
		cs = append(cs, &Constraint{
			Type:             ConstraintTypeForeignKey,
			Columns:          []*Column{c},
			ReferenceColumns: []*Column{c.Reference},
			ReferenceTable:   c.Reference.Table,
			Table:            c.Table,
		})
	}

	return cs
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
