package pqt

import "fmt"

const (
	// RelationshipTypeOneToOne is a relationship that each row in one database table is linked to 1 and only 1 other row in another table.
	// In a one-to-one relationship between Table A and Table B, each row in Table A is linked to another row in Table B.
	// The number of rows in Table A must equal the number of rows in Table B
	RelationshipTypeOneToOne RelationshipType = iota
	// RelationshipTypeOneToMany is a relationship that each row in the related to table can be related to many rows in the relating table.
	// This allows frequently used information to be saved only once in a table and referenced many times in all other tables.
	RelationshipTypeOneToMany
	// RelationshipTypeManyToOne works like one to many, but it points to another owner.
	RelationshipTypeManyToOne

	// RelationshipTypeManyToMany is combination of two many to one relationships.
	// Needs proxy table.
	RelationshipTypeManyToMany

	// NoAction produce an error indicating that the deletion or update would create a foreign key constraint violation.
	// If the constraint is deferred, this error will be produced at constraint check time if there still exist any referencing rows.
	// This is the default action.
	NoAction int32 = iota
	// Restrict produce an error indicating that the deletion or update would create a foreign key constraint violation.
	// This is the same as NO ACTION except that the check is not deferrable.
	Restrict
	// Cascade delete any rows referencing the deleted row,
	// or update the values of the referencing column(s) to the new values of the referenced columns, respectively.
	Cascade
	// SetNull set the referencing column(s) to null
	SetNull
	// SetDefault set the referencing column(s) to their default values.
	// (There must be a row in the referenced table matching the default values, if they are not null, or the operation will fail
	SetDefault
)

// RelationshipType can be used to describe relationship between tables.
// It can be one to one, one to many, many to one or many to many.
type RelationshipType int

// Relationship describes database relationship.
// Usually it is used to describe foreign key constraint.
type Relationship struct {
	// Bidirectional if true means that relationship is bidirectional.
	// It is useful when we want to get all related rows from both tables.
	// If true, the library will generate additional methods for both tables.
	Bidirectional bool
	// Type defines relationship type.
	Type RelationshipType
	// OwnerName is a name of relationship from owner table perspective.
	// For example if we have table A and table B and relationship from A to B,
	// then owner name is a name of relationship from A perspective.
	// It is a good practice to give descriptive names to relationships.
	OwnerName string
	// InversedName is a name of relationship from inversed table perspective.
	// For example if we have table A and table B and relationship from A to B,
	// then inversed name is a name of relationship from B perspective.
	// If not set, the library will generate it automatically.
	// It is useful when we want to have two relationships between two tables or when table self references itself.
	// It is a good practice to give descriptive names to relationships.
	InversedName                            string
	OwnerTable, InversedTable, ThroughTable *Table
	OwnerForeignKey, InversedForeignKey     *Constraint
	OwnerColumns, InversedColumns           Columns
	ColumnName                              string
	OnDelete, OnUpdate                      int32
}

func newRelationship(owner, inversed, through *Table, rt RelationshipType, opts ...RelationshipOption) *Relationship {
	r := &Relationship{
		Type:          rt,
		ThroughTable:  through,
		OwnerTable:    owner,
		InversedTable: inversed,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// OneToOne is a handy constructor that instantiates basic one-to-one relationship.
// It can be adjusted using RelationshipOption.
func OneToOne(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(nil, t, nil, RelationshipTypeOneToOne, opts...)
}

// OneToMany is a handy constructor that instantiates basic one-to-many relationship.
// It can be adjusted using RelationshipOption.
func OneToMany(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(t, nil, nil, RelationshipTypeOneToMany, opts...)
}

// ManyToOne is a handy constructor that instantiates basic many-to-one relationship.
// It can be adjusted using RelationshipOption.
func ManyToOne(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(nil, t, nil, RelationshipTypeManyToOne, opts...)
}

// ManyToMany is a handy constructor that instantiates basic many-to-many relationship.
// It can be adjusted using RelationshipOption.
func ManyToMany(t1 *Table, t2 *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(t1, t2, nil, RelationshipTypeManyToMany, opts...)
}

// RelationshipOption configures how we set up the relationship.
type RelationshipOption func(*Relationship)

// WithOwnerForeignKey ...
func WithOwnerForeignKey(primaryColumns, referenceColumns Columns, opts ...ConstraintOption) RelationshipOption {
	return func(r *Relationship) {
		if r.Type != RelationshipTypeManyToMany {
			panic("function WithOwnerForeignKey can be used only with M2M relationships")
		}

		for _, c := range primaryColumns {
			if c.Table != r.OwnerTable {
				panic(fmt.Sprintf("owner table primary columns inconsistency: column[%v] -> table[%v]", c.Table, r.OwnerTable))
			}
		}

		r.OwnerForeignKey = ForeignKey(primaryColumns, referenceColumns, opts...)
		r.OwnerColumns = primaryColumns
	}
}

// WithInversedForeignKey adjust relationship to have inversed foreign key.
func WithInversedForeignKey(primaryColumns, referenceColumns Columns, opts ...ConstraintOption) RelationshipOption {
	return func(r *Relationship) {
		if r.Type != RelationshipTypeManyToMany {
			panic("function WithInversedForeignKey can be used only with M2M relationships")
		}

		for _, c := range referenceColumns {
			if c.Table != r.InversedTable {
				panic(fmt.Sprintf("inversed table primary columns inconsistency: column[%v] -> table[%v]", c.Table, r.InversedTable))
			}
		}
		r.InversedForeignKey = ForeignKey(primaryColumns, referenceColumns, opts...)
		r.InversedColumns = referenceColumns
	}
}

// WithForeignKey adjust relationship to have foreign key.
func WithForeignKey(primaryColumns, referenceColumns Columns, opts ...ConstraintOption) RelationshipOption {
	return func(r *Relationship) {
		if r.Type == RelationshipTypeManyToMany {
			panic("function WithForeignKey cannot be used with M2M relationships")
		}
		for _, c := range referenceColumns {
			if c.Table != r.InversedTable {
				panic(fmt.Sprintf("inversed table primary columns inconsistency: column[%v] -> table[%v]", c.Table, r.InversedTable))
			}
		}
		r.OwnerForeignKey = ForeignKey(primaryColumns, referenceColumns, opts...)
		r.OwnerColumns = primaryColumns
		r.InversedColumns = referenceColumns
	}
}

// WithInversedName adjust relationship by setting inversed name.
func WithInversedName(s string) RelationshipOption {
	return func(r *Relationship) {
		r.InversedName = s
	}
}

// WithColumnName adjust relationship by setting column name.
func WithColumnName(n string) RelationshipOption {
	return func(r *Relationship) {
		r.ColumnName = n
	}
}

// WithBidirectional adjust relationship by setting bidirectional flag.
func WithBidirectional() RelationshipOption {
	return func(r *Relationship) {
		r.Bidirectional = true
	}
}

// WithOwnerName adjust relationship by setting owner name.
func WithOwnerName(s string) RelationshipOption {
	return func(r *Relationship) {
		r.OwnerName = s
	}
}
