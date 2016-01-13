package pqt

type Table struct {
	Name, Collate, TableSpace string
	IfNotExists, Temporary    bool
	Schema                    *Schema
	Columns                   []*Column
	Constraints               []*Constraint
}

type TableOpts struct {
	Collate, TableSpace    string
	IfNotExists, Temporary bool
}

// NewTable ...
func NewTable(name string) *Table {
	return &Table{
		Name: name,
	}
}

func NewTableWithOpts(name string, opts TableOpts) *Table {
	return &Table{
		Name:        name,
		Temporary:   opts.Temporary,
		Collate:     opts.Collate,
		TableSpace:  opts.TableSpace,
		IfNotExists: opts.IfNotExists,
	}
}

func (t *Table) FullName() string {
	if t.Schema != nil && t.Schema.Name != "" {
		return t.Schema.Name + "." + t.Name
	}

	return t.Name
}

func (t *Table) AddColumn(c *Column) *Table {
	if t.Columns == nil {
		t.Columns = make([]*Column, 0, 1)
	}

	if c.Table == nil {
		c.Table = t
	} else {
		*c.Table = *t
	}
	t.Columns = append(t.Columns, c)

	return t
}

func (t *Table) AddConstraint(c *Constraint) *Table {
	if t.Constraints == nil {
		t.Constraints = make([]*Constraint, 0, 1)
	}

	if c.Table == nil {
		c.Table = t
	} else {
		*c.Table = *t
	}

	t.Constraints = append(t.Constraints, c)

	return t
}

func (t *Table) AddCheck(check string, columns ...*Column) *Table {
	return t.AddConstraint(Check(t, check, columns...))
}

func (t *Table) SetIfNotExists(ine bool) *Table {
	t.IfNotExists = ine
	return t
}

func (t *Table) SetSchema(s *Schema) *Table {
	if t.Schema == nil {
		t.Schema = s
	} else {
		*t.Schema = *s
	}

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
