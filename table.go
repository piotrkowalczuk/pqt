package pqt

import (
	"fmt"
	"sort"
)

// Table is partially implemented postgres table synopsis.
type Table struct {
	self                                 bool
	Name, ShortName, Collate, TableSpace string
	IfNotExists, Temporary               bool
	// Schema references parent schema.
	Schema *Schema
	// Columns is a collection of columns that table contains.
	Columns Columns
	// Constraints is a collection of constraints that table contains.
	Constraints Constraints
	// OwnedRelationships is a collection of relationships that table owns.
	OwnedRelationships []*Relationship
	// InversedRelationships is a collection of relationships that table is inversed in.
	InversedRelationships []*Relationship
	// ManyToManyRelationships is a collection of relationships that table is inversed in.
	ManyToManyRelationships []*Relationship
}

// NewTable allocates new table using given name and options.
func NewTable(name string, opts ...TableOption) *Table {
	t := &Table{
		Name:                  name,
		ShortName:             name,
		Columns:               make(Columns, 0),
		Constraints:           make([]*Constraint, 0),
		InversedRelationships: make([]*Relationship, 0),
		OwnedRelationships:    make([]*Relationship, 0),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// SelfReference returns almost empty table that express self reference.
// Should be used with relationships.
func SelfReference() *Table {
	return &Table{
		self: true,
	}
}

// FullName if schema is defined returns name in format <schema>.<name> or just <name> if not set.
func (t *Table) FullName() string {
	if t.Schema != nil && t.Schema.Name != "" {
		return t.Schema.Name + "." + t.Name
	}

	return t.Name
}

// AddColumn adds column to the table.
func (t *Table) AddColumn(c *Column) *Table {
	if c.Reference != nil {
		r := newRelationship(t, c.Reference.Table, nil, RelationshipTypeManyToOne, c.ReferenceOptions...)
		r.OwnerColumns = Columns{c}
		r.InversedColumns = Columns{c.Reference}

		t.OwnedRelationships = append(t.OwnedRelationships, r)
		if r.Bidirectional && r.InversedTable != nil {
			r.InversedTable.InversedRelationships = append(r.InversedTable.InversedRelationships, r)
		}

		t.AddConstraint(&Constraint{
			Type:           ConstraintTypeForeignKey,
			Table:          r.InversedTable,
			Columns:        r.InversedColumns,
			PrimaryTable:   r.OwnerTable,
			PrimaryColumns: r.OwnerColumns,
			OnDelete:       c.OnDelete,
			OnUpdate:       c.OnUpdate,
			Match:          c.Match,
		})
		// When constraint is created, redundant data from column needs to be removed.
		c.Reference = nil
		c.OnDelete = 0
		c.OnUpdate = 0
		c.Match = 0
	}

	return t.addColumn(c)
}

func (t *Table) addColumn(c *Column) *Table {
	if t.Columns == nil {
		t.Columns = make(Columns, 0, 1)
	}

	c.Table = t
	t.Columns = append(t.Columns, c)

	t.Constraints = append(t.Constraints, c.Constraints()...)
	sort.Sort(&t.Columns)
	return t
}

// AddRelationship adds relationship to the table.
func (t *Table) AddRelationship(r *Relationship, opts ...ColumnOption) *Table {
	if r == nil {
		return t
	}
	if r.Type == RelationshipTypeManyToMany {
		return t.addRelationshipManyToMany(r, opts...)
	}
	return t.addRelationship(r, opts...)
}

func (t *Table) addRelationship(r *Relationship, opts ...ColumnOption) *Table {
	switch {
	case r.InversedTable != nil && r.InversedTable.self:
		r.InversedTable = t
	case r.OwnerTable != nil && r.OwnerTable.self:
		r.OwnerTable = t
	}

	switch r.Type {
	case RelationshipTypeOneToOne, RelationshipTypeManyToOne:
		r.OwnerTable = t
	case RelationshipTypeOneToMany:
		r.InversedTable = t
	}

	r.OwnerTable.OwnedRelationships = append(r.OwnerTable.OwnedRelationships, r)
	if r.Bidirectional {
		r.InversedTable.InversedRelationships = append(r.InversedTable.InversedRelationships, r)
	}

	pk, ok := r.InversedTable.PrimaryKey()
	if !ok {
		return t
	}

	name := r.ColumnName
	if name == "" {
		name = r.InversedTable.Name + "_" + pk.Name
	}

	nt := fkType(pk.Type)

	col := NewColumn(name, nt, append([]ColumnOption{WithReference(pk)}, opts...)...)
	r.OwnerTable.addColumn(col)
	r.OwnerColumns = Columns{col}
	r.InversedColumns = Columns{pk}

	return t
}

func (t *Table) addRelationshipManyToMany(r *Relationship, opts ...ColumnOption) *Table {
	r.ThroughTable = t
	r.ThroughTable.OwnedRelationships = append(r.ThroughTable.OwnedRelationships, r)

	if r.Bidirectional {
		r.OwnerTable.ManyToManyRelationships = append(r.OwnerTable.ManyToManyRelationships, r)
		r.InversedTable.ManyToManyRelationships = append(r.InversedTable.ManyToManyRelationships, r)
	}

	ownerColumns := make(Columns, 0)
	inversedColumns := make(Columns, 0)

	if r.OwnerForeignKey != nil {
		r.OwnerForeignKey.PrimaryTable = r.ThroughTable
		r.ThroughTable.AddConstraint(r.OwnerForeignKey)
		for _, oc := range r.OwnerForeignKey.PrimaryColumns {
			r.ThroughTable.AddColumn(oc)
			ownerColumns = append(ownerColumns, oc)
		}
	} else {
		pk, ok := r.OwnerTable.PrimaryKey()
		if !ok {
			panic(fmt.Sprintf("missing owner table (%s) primary key for many to many relationship", r.OwnerTable.Name))
		}

		name := r.ColumnName
		if name == "" {
			name = r.OwnerTable.Name + "_" + pk.Name
		}

		nt := fkType(pk.Type)

		oc := NewColumn(name, nt, append([]ColumnOption{WithReference(pk)}, opts...)...)
		r.ThroughTable.AddColumn(oc)
		ownerColumns = append(ownerColumns, oc)
	}

	if r.InversedForeignKey != nil {
		r.InversedForeignKey.PrimaryTable = r.ThroughTable
		r.ThroughTable.AddConstraint(r.InversedForeignKey)
		for _, ic := range r.InversedForeignKey.PrimaryColumns {
			r.ThroughTable.AddColumn(ic)
			ownerColumns = append(ownerColumns, ic)
		}
	} else {
		pk, ok := r.InversedTable.PrimaryKey()
		if !ok {
			panic(fmt.Sprintf("missing inversed table (%s) primary key for many to many relationship", r.InversedTable.Name))
		}
		name := r.ColumnName
		if name == "" {
			name = r.InversedTable.Name + "_" + pk.Name
		}
		nt := fkType(pk.Type)
		ic := NewColumn(name, nt, append([]ColumnOption{WithReference(pk)}, opts...)...)
		r.ThroughTable.AddColumn(ic)
		inversedColumns = append(inversedColumns, ic)
	}

	r.ThroughTable.AddUnique(append(ownerColumns, inversedColumns...)...)
	return t
}

// AddConstraint adds constraint to the table.
func (t *Table) AddConstraint(c *Constraint) *Table {
	if t.Constraints == nil {
		t.Constraints = make([]*Constraint, 0, 1)
	}

	c.PrimaryTable = t

	t.Constraints = append(t.Constraints, c)

	return t
}

// AddCheck adds check constraint to the table.
func (t *Table) AddCheck(check string, columns ...*Column) *Table {
	return t.AddConstraint(Check(t, check, columns...))
}

// AddUnique adds unique constraint to the table.
func (t *Table) AddUnique(columns ...*Column) *Table {
	return t.AddConstraint(Unique(t, columns...))
}

// AddIndex adds index to the table.
func (t *Table) AddIndex(columns ...*Column) *Table {
	return t.AddConstraint(Index(t, columns...))
}

// AddUniqueIndex ...
func (t *Table) AddUniqueIndex(methodSuffix, where string, columns ...*Column) *Table {
	return t.AddConstraint(UniqueIndex(t, methodSuffix, where, columns...))
}

// SetIfNotExists sets IfNotExists flag.
func (t *Table) SetIfNotExists(ine bool) *Table {
	t.IfNotExists = ine
	return t
}

// SetSchema sets schema name table belongs to.
func (t *Table) SetSchema(s *Schema) *Table {
	t.Schema = s

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

// TableOption configures how we set up the table.
type TableOption func(*Table)

// WithTableIfNotExists is table option that sets IfNotExists flag to true.
func WithTableIfNotExists() TableOption {
	return func(t *Table) {
		t.IfNotExists = true
	}
}

// WithTemporary specified, the table is created as a temporary table.
// Temporary tables are automatically dropped at the end of a session, or optionally at the end of the current transaction (see ON COMMIT below).
// Existing permanent tables with the same name are not visible to the current session while the temporary table exists, unless they are referenced with schema-qualified names.
// Any indexes created on a temporary table are automatically temporary as well.
func WithTemporary() TableOption {
	return func(t *Table) {
		t.Temporary = true
	}
}

// WithTableSpace pass the name of the tablespace in which the new table is to be created.
// If not specified, default_tablespace is consulted, or temp_tablespaces if the table is temporary.
func WithTableSpace(s string) TableOption {
	return func(t *Table) {
		t.TableSpace = s
	}
}

// WithTableShortName pass the short name of the table.
func WithTableShortName(s string) TableOption {
	return func(t *Table) {
		t.ShortName = s
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
