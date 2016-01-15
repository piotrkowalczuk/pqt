package pqt

// Schema ...
type Schema struct {
	Name      string
	Tables    []*Table
	Types     []Type
	Functions []*Function
}

// NewSchema ...
func NewSchema(name string) *Schema {
	return &Schema{
		Name: name,
	}
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
