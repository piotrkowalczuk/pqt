package pqt

// Schema ...
type Schema struct {
	Name        string
	IfNotExists bool
	Tables      []*Table
	Functions   []*Function
	Types       []Type
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

func (s *Schema) AddFunction(f *Function) *Schema {
	if s.Functions == nil {
		s.Functions = make([]*Function, 0, 1)
	}
	s.Functions = append(s.Functions, f)

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
