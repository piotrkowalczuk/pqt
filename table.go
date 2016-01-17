package pqt

// Table ...
type Table struct {
	self                      bool
	Name, Collate, TableSpace string
	IfNotExists, Temporary    bool
	Schema                    *Schema
	Columns                   []*Column
	Constraints               []*Constraint
	OwnedRelationships        []*Relationship
	InversedRelationships     []*Relationship
	ManyToManyRelationships   []*Relationship
}

// TableOption configures how we set up the table.
type TableOption func(*Table)

// NewTable ...
func NewTable(name string, opts ...TableOption) *Table {
	t := &Table{
		Name:                  name,
		Columns:               make([]*Column, 0),
		Constraints:           make([]*Constraint, 0),
		InversedRelationships: make([]*Relationship, 0),
		OwnedRelationships:    make([]*Relationship, 0),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// SelfRefference returns almost empty table that express self reference. Should be used with relationships.
func SelfRefference() *Table {
	return &Table{
		self: true,
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
func (t *Table) AddRelationship(r *Relationship, opts ...ColumnOption) *Table {
	if r == nil {
		return t
	}

	if r.Type == RelationshipTypeManyToMany {
		return t.addRelationshipManyToMany(r, opts...)
	}
	if r.InversedTable != nil && r.InversedTable.self {
		r.InversedTable = t
	}

	switch r.Type {
	case RelationshipTypeOneToOne, RelationshipTypeManyToOne:
		r.OwnerTable = t
	case RelationshipTypeOneToMany:
		r.InversedTable = t
	}
	pk, ok := r.InversedTable.PrimaryKey()
	if !ok {
		return t
	}

	name := r.ColumnName
	if name == "" {
		name = t.Name + "_" + pk.Name
	}

	nt := fkType(pk.Type)

	switch r.Type {
	case RelationshipTypeOneToOne, RelationshipTypeManyToOne:
		r.OwnerTable.OwnedRelationships = append(r.OwnerTable.OwnedRelationships, r)
		if r.Bidirectional {
			r.InversedTable.InversedRelationships = append(r.InversedTable.InversedRelationships, r)
		}
	case RelationshipTypeOneToMany:
		r.OwnerTable.OwnedRelationships = append(r.OwnerTable.OwnedRelationships, r)
		if r.Bidirectional {
			r.InversedTable.InversedRelationships = append(r.InversedTable.InversedRelationships, r)
		}
	}

	r.OwnerTable.AddColumn(NewColumn(name, nt, append([]ColumnOption{WithReference(pk)}, opts...)...))

	return t
}

func (t *Table) addRelationshipManyToMany(r *Relationship, opts ...ColumnOption) *Table {
	r.ThroughTable = t
	r.ThroughTable.OwnedRelationships = append(r.ThroughTable.OwnedRelationships, r)

	if r.Bidirectional {
		r.InversedTable.ManyToManyRelationships = append(r.InversedTable.ManyToManyRelationships, r)
		r.OwnerTable.ManyToManyRelationships = append(r.OwnerTable.ManyToManyRelationships, r)
	}

	pk1, ok := r.OwnerTable.PrimaryKey()
	if !ok {
		return t
	}
	pk2, ok := r.InversedTable.PrimaryKey()
	if !ok {
		return t
	}

	name1 := r.ColumnName
	if name1 == "" {
		name1 = r.OwnerTable.Name + "_" + pk1.Name
	}

	name2 := r.ColumnName
	if name2 == "" {
		name2 = r.InversedTable.Name + "_" + pk2.Name
	}

	nt1 := fkType(pk1.Type)
	nt2 := fkType(pk2.Type)
	r.ThroughTable.AddColumn(NewColumn(name1, nt1, append([]ColumnOption{WithReference(pk1)}, opts...)...))
	r.ThroughTable.AddColumn(NewColumn(name2, nt2, append([]ColumnOption{WithReference(pk2)}, opts...)...))

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

// AddUnique ...
func (t *Table) AddUnique(columns ...*Column) *Table {
	return t.AddConstraint(Unique(t, columns...))
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

// WithIfNotExists ...
func WithIfNotExists() func(*Table) {
	return func(t *Table) {
		t.IfNotExists = true
	}
}

// WithTemporary ...
func WithTemporary() func(*Table) {
	return func(t *Table) {
		t.Temporary = true
	}
}

// WithTableSpace ...
func WithTableSpace(s string) func(*Table) {
	return func(t *Table) {
		t.TableSpace = s
	}
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
