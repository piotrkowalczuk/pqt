package pqt

// Schema describes database schema.
// It is a collection of tables, functions and types.
type Schema struct {
	// Name is a schema name.
	Name string
	// IfNotExists if true means that schema should be created only if it does not exist.
	// If true, creation process will not fail if schema already exists.
	IfNotExists bool
	// Tables is a collection of tables that schema contains.
	Tables []*Table
	// Functions is a collection of functions that schema contains.
	Functions []*Function
	// Types is a collection of types that schema contains.
	Types []Type
}

// NewSchema initializes new instance of Schema for given name and options.
func NewSchema(name string, opts ...SchemaOption) *Schema {
	s := &Schema{
		Name: name,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddTable adds table to schema.
func (s *Schema) AddTable(t *Table) *Schema {
	if s.Tables == nil {
		s.Tables = make([]*Table, 0, 1)
	}

	t.Schema = s
	s.Tables = append(s.Tables, t)
	return s
}

// AddFunction adds function to schema.
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
