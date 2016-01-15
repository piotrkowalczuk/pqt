package pqt

const (
	RelationshipTypeOneToOneUnidirectional RelationshipType = iota
	RelationshipTypeOneToOneBidirectional
	RelationshipTypeOneToOneSelfReferencing
	RelationshipTypeOneToMany
	RelationshipTypeOneToManySelfReferencing
	RelationshipTypeManyToMany
	RelationshipTypeManyToManySelfReferencing

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
	Optional                              bool
	Type                                  RelationshipType
	InversedColumnName                    string
	MappedBy, InversedBy                  string
	MappedTable, InversedTable, JoinTable *Table
	OnDelete, OnUpdate                    int32
}

func newRelationship(t, j *Table, rt RelationshipType, opts ...relationshipOpt) *Relationship {
	r := &Relationship{
		Type:        rt,
		MappedTable: t,
		JoinTable:   j,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// OneToOneUnidirectional ...
func OneToOneUnidirectional(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToOneUnidirectional, opts...)
}

// OneToOneBidirectional ...
func OneToOneBidirectional(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToOneBidirectional, opts...)
}

// OneToOneSelfReferencing ...
func OneToOneSelfReferencing(opts ...relationshipOpt) *Relationship {
	return newRelationship(nil, nil, RelationshipTypeOneToOneSelfReferencing, opts...)
}

// OneToMany ...
func OneToMany(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToMany, opts...)
}

// OneToManySelfReferencing ...
func OneToManySelfReferencing(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToManySelfReferencing, opts...)
}

// ManyToMany ...
func ManyToMany(t, j *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, j, RelationshipTypeManyToMany, opts...)
}

// ManyToManySelfReferencing ...
func ManyToManySelfReferencing(j *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(nil, j, RelationshipTypeManyToManySelfReferencing, opts...)
}

type relationshipOpt func(*Relationship)

// WithMappedBy ...
func WithMappedBy(s string) relationshipOpt {
	return func(r *Relationship) {
		r.MappedBy = s
	}
}

// WithInversedBy ...
func WithInversedBy(s string) relationshipOpt {
	return func(r *Relationship) {
		r.InversedBy = s
	}
}

// WithInversedColumnName ...
func WithInversedColumnName(n string) relationshipOpt {
	return func(r *Relationship) {
		r.InversedColumnName = n
	}
}

func WithOptional() relationshipOpt {
	return func(r *Relationship) {
		r.Optional = true
	}
}

// WithOnDelete add ON DELETE clause that specifies the action to perform when a referenced row in the referenced table is being deleted
func WithOnDelete(on int32) relationshipOpt {
	return func(r *Relationship) {
		r.OnDelete = on
	}
}

// WithOnUpdate add ON UPDATE clause that specifies the action to perform when a referenced column in the referenced table is being updated to a new value.
func WithOnUpdate(on int32) relationshipOpt {
	return func(r *Relationship) {
		r.OnUpdate = on
	}
}
