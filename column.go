package pqt

import (
	"bytes"
	"strings"
)

const (
	// EventInsert ...
	EventInsert Event = "INSERT"
	// EventUpdate ...
	EventUpdate Event = "UPDATE"
)

// Event ...
type Event string

// Column ...
type Column struct {
	Name, ShortName, Collate, Check                                      string
	Default                                                              map[Event]string
	NotNull, Unique, PrimaryKey, Index                                   bool
	Type                                                                 Type
	Table                                                                *Table
	Reference                                                            *Column
	ReferenceOptions                                                     []RelationshipOption
	Match, OnDelete, OnUpdate                                            int32
	NoInherit, DeferrableInitiallyDeferred, DeferrableInitiallyImmediate bool
	// Dynamic
	IsDynamic bool
	Func      *Function
	Columns   Columns
}

// NewColumn ...
func NewColumn(n string, t Type, opts ...ColumnOption) *Column {
	c := &Column{
		Name: n,
		Type: t,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewDynamicColumn ...
func NewDynamicColumn(n string, f *Function, cs ...*Column) *Column {
	return &Column{
		IsDynamic: true,
		NotNull:   true,
		Name:      n,
		Type:      f.Type,
		Func:      f,
		Columns:   cs,
	}
}

// Constraints ...
func (c *Column) Constraints() []*Constraint {
	var cs []*Constraint

	if c.PrimaryKey {
		cs = append(cs, &Constraint{
			Type:           ConstraintTypePrimaryKey,
			PrimaryColumns: Columns{c},
			PrimaryTable:   c.Table,
		})
	} else if c.Unique {
		cs = append(cs, &Constraint{
			Type:           ConstraintTypeUnique,
			PrimaryColumns: Columns{c},
			PrimaryTable:   c.Table,
		})
	} else if c.Index {
		cs = append(cs, &Constraint{
			Type:           ConstraintTypeIndex,
			PrimaryColumns: Columns{c},
			PrimaryTable:   c.Table,
		})
	}
	if c.Check != "" {
		cs = append(cs, &Constraint{
			Type:           ConstraintTypeCheck,
			Check:          c.Check,
			PrimaryColumns: Columns{c},
			PrimaryTable:   c.Table,
		})
	}
	if c.Reference != nil {
		cs = append(cs, &Constraint{
			Type:           ConstraintTypeForeignKey,
			PrimaryColumns: Columns{c},
			Columns:        Columns{c.Reference},
			Table:          c.Reference.Table,
			PrimaryTable:   c.Table,
		})
	}

	return cs
}

// DefaultOn ...
func (c Column) DefaultOn(e ...Event) (string, bool) {
	for k, v := range c.Default {
		for _, ee := range e {
			if k == ee {
				return v, true
			}
		}
	}

	return "", false
}

// Columns is a slice of columns that implements few handy methods.
type Columns []*Column

// Len implements sort.Interface interface.
func (c Columns) Len() int {
	return len(c)
}

// Swap implements sort.Interface interface.
func (c Columns) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less implements sort.Interface interface.
func (c Columns) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

// String implements Stringer interface.
func (c Columns) String() string {
	b := bytes.NewBuffer(nil)
	defer b.Reset()

	for i, col := range c {
		if i != 0 {
			b.WriteRune(',')
		}
		b.WriteString(col.Name)
	}

	return b.String()
}

// JoinColumns ...
func JoinColumns(columns Columns, sep string) string {
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
	Schema                        *Schema
	Type                          Type
}

// Constraint ...
func (a *Attribute) Constraint() (*Constraint, bool) {
	var kind ConstraintType
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

// ColumnOption configures how we set up the column.
type ColumnOption func(*Column)

// WithTypeMapping ...
func WithTypeMapping(t Type) ColumnOption {
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
func WithCheck(ch string) ColumnOption {
	return func(c *Column) {
		c.Check = ch
	}
}

// WithUnique ...
func WithUnique() ColumnOption {
	return func(c *Column) {
		c.Unique = true
	}
}

// WithIndex ...
func WithIndex() ColumnOption {
	return func(c *Column) {
		c.Index = true
	}
}

// WithPrimaryKey ...
func WithPrimaryKey() ColumnOption {
	return func(c *Column) {
		c.PrimaryKey = true
	}
}

// WithCollate ...
func WithCollate(cl string) ColumnOption {
	return func(c *Column) {
		c.Collate = cl
	}
}

// WithDefault ...
func WithDefault(d string, e ...Event) ColumnOption {
	return func(c *Column) {
		if len(e) == 0 {
			e = []Event{EventInsert}
		}

		if c.Default == nil {
			c.Default = make(map[Event]string, len(e))
		}

		for _, event := range e {
			c.Default[event] = d
		}
	}
}

// WithNotNull ...
func WithNotNull() ColumnOption {
	return func(c *Column) {
		c.NotNull = true
	}
}

// WithReference ...
func WithReference(r *Column, opts ...RelationshipOption) ColumnOption {
	return func(c *Column) {
		c.Reference = r
		c.ReferenceOptions = opts
	}
}

// WithOnDelete add ON DELETE clause that specifies the action to perform when a referenced row in the referenced table is being deleted
func WithOnDelete(on int32) ColumnOption {
	return func(c *Column) {
		c.OnDelete = on
	}
}

// WithOnUpdate add ON UPDATE clause that specifies the action to perform when a referenced column in the referenced table is being updated to a new value.
func WithOnUpdate(on int32) ColumnOption {
	return func(c *Column) {
		c.OnUpdate = on
	}
}

// WithColumnShortName ...
func WithColumnShortName(s string) ColumnOption {
	return func(c *Column) {
		c.ShortName = s
	}
}
