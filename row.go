package pqt

import "strings"

// Column ...
type Column struct {
	Name, Collate, Default, Check string
	NotNull, Unique, PrimaryKey   bool
	Type                          Type
	Table                         *Table
	Reference                     *Column
}

// NewColumn ...
func NewColumn(n string, t Type, opts ...columnOpt) *Column {
	c := &Column{
		Name: n,
		Type: t,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Constraints ...
func (c *Column) Constraints() []*Constraint {
	var cs []*Constraint

	if c.PrimaryKey {
		cs = append(cs, &Constraint{
			Type:    ConstraintTypePrimaryKey,
			Columns: []*Column{c},
			Table:   c.Table,
		})
	} else if c.Unique {
		cs = append(cs, &Constraint{
			Type:    ConstraintTypeUnique,
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

// JoinColumns ...
func JoinColumns(columns []*Column, sep string) string {
	tmp := make([]string, 0, len(columns))
	for _, c := range columns {
		tmp = append(tmp, c.Name)
	}

	return strings.Join(tmp, sep)
}

// Attribute ...
type Attribute struct {
	Name, Collate, Default, Check string
	NotNull, Unique, PrimaryKey   bool
	Type                          Type
}

// Constraint ...
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

type columnOpt func(*Column)

// WithType ...
func WithType(t Type) columnOpt {
	return func(c *Column) {
		c.Type = t
	}
}

// WithTypeMapping ...
func WithTypeMapping(t Type) columnOpt {
	return func(c *Column) {
		switch ct := c.Type.(type) {
		case MappableType:
			ct.Mapping = append(ct.Mapping, t)
		default:
			c.Type = TypeMappable(c.Type, t)
		}
	}
}

// WithCheck ...
func WithCheck(ch string) columnOpt {
	return func(c *Column) {
		c.Check = ch
	}
}

// WithUnique ...
func WithUnique() columnOpt {
	return func(c *Column) {
		c.Unique = true
	}
}

// WithPrimaryKey ...
func WithPrimaryKey() columnOpt {
	return func(c *Column) {
		c.PrimaryKey = true
	}
}

// WithCollate ...
func WithCollate(cl string) columnOpt {
	return func(c *Column) {
		c.Collate = cl
	}
}

// WithDefault ...
func WithDefault(d string) columnOpt {
	return func(c *Column) {
		c.Default = d
	}
}

// WithNotNull ...
func WithNotNull() columnOpt {
	return func(c *Column) {
		c.NotNull = true
	}
}

// WithReference ...
func WithReference(r *Column) columnOpt {
	return func(c *Column) {
		c.Reference = r
	}
}
