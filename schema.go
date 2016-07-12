package pqt

// Schema ...
type Schema struct {
	Name        string
	IfNotExists bool
	Tables      []*Table
	Types       []Type
	Functions   []*Function
}

// NewSchema ...
func NewSchema(name string, opts ...SchemaOption) *Schema {
	s := &Schema{
		Name: name,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddTable ...
func (s *Schema) AddTable(t *Table) *Schema {
	if s.Tables == nil {
		s.Tables = make([]*Table, 0, 1)
	}

	if t.Schema == nil {
		t.Schema = s
	} else {
		*t.Schema = *s
	}
	s.Tables = append(s.Tables, t)
	return s
}

// SchemaOption configures how we set up a schema.
type SchemaOption func(*Schema)

// WithSchemaIfNotExists is schema option that sets IfNotExists flag to true.
func WithSchemaIfNotExists() SchemaOption {
	return func(s *Schema) {
		s.IfNotExists = true
	}
}
