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

type RelationshipType int

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

func OneToOneUnidirectional(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToOneUnidirectional, opts...)
}

func OneToOneBidirectional(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToOneBidirectional, opts...)
}

func OneToOneSelfReferencing(opts ...relationshipOpt) *Relationship {
	return newRelationship(nil, nil, RelationshipTypeOneToOneSelfReferencing, opts...)
}

func OneToMany(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToMany, opts...)
}

func OneToManySelfReferencing(t *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, nil, RelationshipTypeOneToManySelfReferencing, opts...)
}

func ManyToMany(t, j *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(t, j, RelationshipTypeManyToMany, opts...)
}

func ManyToManySelfReferencing(j *Table, opts ...relationshipOpt) *Relationship {
	return newRelationship(nil, j, RelationshipTypeManyToManySelfReferencing, opts...)
}

type relationshipOpt func(*Relationship)

func WithMappedBy(s string) relationshipOpt {
	return func(r *Relationship) {
		r.MappedBy = s
	}
}

func WithInversedBy(s string) relationshipOpt {
	return func(r *Relationship) {
		r.InversedBy = s
	}
}
