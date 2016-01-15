package pqt

const (
	RelationshipTypeOneToOneUnidirectional RelationshipType = iota
	RelationshipTypeOneToOneBidirectional
	RelationshipTypeOneToOneSelfReferencing
	RelationshipTypeOneToMany
	RelationshipTypeOneToManySelfReferencing
	RelationshipTypeManyToMany
	RelationshipTypeManyToManySelfReferencing
)

// RelationshipType ...
type RelationshipType int

// Relationship ...
type Relationship struct {
	Type                                  RelationshipType
	MappedBy, InversedBy                  string
	MappedTable, InversedTable, JoinTable *Table
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
