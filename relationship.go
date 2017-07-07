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

// RelationshipType ...
type RelationshipType int

// Relationship ...
type Relationship struct {
	Bidirectional                           bool
	Type                                    RelationshipType
	OwnerName, InversedName                 string
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

// OneToOne ...
func OneToOne(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(nil, t, nil, RelationshipTypeOneToOne, opts...)
}

// OneToMany ...
func OneToMany(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(t, nil, nil, RelationshipTypeOneToMany, opts...)
}

// ManyToOne ...
func ManyToOne(t *Table, opts ...RelationshipOption) *Relationship {
	return newRelationship(nil, t, nil, RelationshipTypeManyToOne, opts...)
}

// ManyToMany ...
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

// WithInversedForeignKey ...
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

// WithForeignKey ...
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

// WithInversedName ...
func WithInversedName(s string) RelationshipOption {
	return func(r *Relationship) {
		r.InversedName = s
	}
}

// WithColumnName ...
func WithColumnName(n string) RelationshipOption {
	return func(r *Relationship) {
		r.ColumnName = n
	}
}

// WithBidirectional ...
func WithBidirectional() RelationshipOption {
	return func(r *Relationship) {
		r.Bidirectional = true
	}
}

// WithOwnerName ...
func WithOwnerName(s string) RelationshipOption {
	return func(r *Relationship) {
		r.OwnerName = s
	}
}
