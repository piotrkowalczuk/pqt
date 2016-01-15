package pqt

// Table ...
type Table struct {
	self                      bool
	Name, Collate, TableSpace string
	IfNotExists, Temporary    bool
	Schema                    *Schema
	Columns                   []*Column
	Constraints               []*Constraint
	Relationships             []*Relationship
}

// TableOpts ...
type TableOpts struct {
	Collate, TableSpace    string
	IfNotExists, Temporary bool
}

// NewTable ...
func NewTable(name string) *Table {
	return &Table{
		Name:          name,
		Columns:       make([]*Column, 0),
		Constraints:   make([]*Constraint, 0),
		Relationships: make([]*Relationship, 0),
	}
}

// NewTableWithOpts ...
func NewTableWithOpts(name string, opts TableOpts) *Table {
	return &Table{
		Name:        name,
		Temporary:   opts.Temporary,
		Collate:     opts.Collate,
		TableSpace:  opts.TableSpace,
		IfNotExists: opts.IfNotExists,
	}
}

// FullName ...
func (t *Table) FullName() string {
	if t.Schema != nil && t.Schema.Name != "" {
		return t.Schema.Name + "." + t.Name
	}

	return t.Name
}

// AddColumn ...
func (t *Table) AddColumn(c *Column) *Table {
	if t.Columns == nil {
		t.Columns = make([]*Column, 0, 1)
	}

	if c.Table == nil {
		c.Table = t
	} else {
		*c.Table = *t
	}
	t.Columns = append(t.Columns, c)

	return t
}

// AddRelationship ...
func (t *Table) AddRelationship(r *Relationship) *Table {
	if r == nil {
		return t
	}

	if r.Type == RelationshipTypeOneToOneSelfReferencing || r.Type == RelationshipTypeOneToManySelfReferencing || r.Type == RelationshipTypeManyToManySelfReferencing {
		r.MappedTable = t
	}

	pk, ok := t.PrimaryKey()
	if !ok {
		return t
	}

	name := t.Name + "_" + pk.Name
	nt := fkType(pk.Type)

	switch r.Type {
	case RelationshipTypeOneToOneBidirectional, RelationshipTypeOneToOneSelfReferencing:
		t.Relationships = append(t.Relationships, &Relationship{
			MappedBy:    r.MappedBy,
			MappedTable: r.MappedTable,
			Type:        r.Type,
		})

		r.MappedTable.Relationships = append(r.MappedTable.Relationships, &Relationship{
			InversedBy:    r.InversedBy,
			InversedTable: t,
			Type:          r.Type,
		})
		r.MappedTable.AddColumn(NewColumn(name, nt, WithUnique(), WithReference(pk)))
	case RelationshipTypeOneToOneUnidirectional:
		r.MappedTable.Relationships = append(r.MappedTable.Relationships, &Relationship{
			InversedBy:    r.InversedBy,
			InversedTable: t,
			Type:          r.Type,
		})
		r.MappedTable.AddColumn(NewColumn(name, nt, WithUnique(), WithReference(pk)))
	case RelationshipTypeOneToMany, RelationshipTypeOneToManySelfReferencing:
		t.Relationships = append(t.Relationships, &Relationship{
			MappedBy:    r.MappedBy,
			MappedTable: r.MappedTable,
			Type:        r.Type,
		})

		r.MappedTable.Relationships = append(r.MappedTable.Relationships, &Relationship{
			InversedBy:    r.InversedBy,
			InversedTable: t,
			Type:          r.Type,
		})
		r.MappedTable.AddColumn(NewColumn(name, nt, WithReference(pk)))
	case RelationshipTypeManyToMany, RelationshipTypeManyToManySelfReferencing:
		t.Relationships = append(t.Relationships, &Relationship{
			MappedBy:    r.MappedBy,
			MappedTable: r.MappedTable,
			Type:        r.Type,
		})

		r.MappedTable.Relationships = append(r.MappedTable.Relationships, &Relationship{
			InversedBy:    r.InversedBy,
			InversedTable: t,
			Type:          r.Type,
		})

		pk2, ok := r.MappedTable.PrimaryKey()
		if !ok {
			return t
		}

		r.JoinTable.AddColumn(NewColumn(t.Name+"_"+pk.Name, fkType(pk.Type), WithNotNull(), WithReference(pk)))
		r.JoinTable.AddColumn(NewColumn(r.MappedTable.Name+"_"+pk2.Name, fkType(pk2.Type), WithNotNull(), WithReference(pk2)))
	}

	return t
}

// AddConstraint ...
func (t *Table) AddConstraint(c *Constraint) *Table {
	if t.Constraints == nil {
		t.Constraints = make([]*Constraint, 0, 1)
	}

	if c.Table == nil {
		c.Table = t
	} else {
		*c.Table = *t
	}

	t.Constraints = append(t.Constraints, c)

	return t
}

// AddCheck ...
func (t *Table) AddCheck(check string, columns ...*Column) *Table {
	return t.AddConstraint(Check(t, check, columns...))
}

// SetIfNotExists ...
func (t *Table) SetIfNotExists(ine bool) *Table {
	t.IfNotExists = ine
	return t
}

// SetSchema ...
func (t *Table) SetSchema(s *Schema) *Table {
	if t.Schema == nil {
		t.Schema = s
	} else {
		*t.Schema = *s
	}

	return t
}

// PrimaryKey returns column that is primary key, or false if none.
func (t *Table) PrimaryKey() (*Column, bool) {
	for _, c := range t.Columns {
		if c.PrimaryKey {
			return c, true
		}
	}

	return nil, false
}

func fkType(t Type) Type {
	switch t {
	case TypeSerial():
		return TypeInteger()
	case TypeSerialBig():
		return TypeIntegerBig()
	case TypeSerialSmall():
		return TypeIntegerSmall()
	default:
		return t
	}
}
