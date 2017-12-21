package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_RepositoryInsert(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").AddRelationship(pqt.ManyToOne(t1))

	g := &gogen.Generator{}
	g.Repository(t2)
	g.RepositoryInsert(t2)
	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) Insert(ctx context.Context, e *T2Entity) (*T2Entity, error) {
	query, args, err := r.InsertQuery(e, true)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, query, args...).Scan()
	if r.Log != nil {
		r.Log(err, TableT2, "insert", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}`)
}

func TestGenerator_RepositoryInsertQuery(t *testing.T) {
	name := pqt.NewColumn("name", pqt.TypeText(), pqt.WithNotNull(), pqt.WithIndex())
	description := pqt.NewColumn("description", pqt.TypeText(), pqt.WithColumnShortName("desc"))

	t1ID := pqt.NewColumn("id", pqt.TypeIntegerBig(), pqt.WithPrimaryKey())

	t1 := pqt.NewTable("t1").AddColumn(t1ID)
	t2 := pqt.NewTable("t2").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig())).
		AddColumn(pqt.NewColumn("t1_id", pqt.TypeIntegerBig(), pqt.WithReference(t1ID))).
		AddColumn(pqt.NewDynamicColumn("slug", &pqt.Function{
			Name:      "slugify",
			Type:      pqt.TypeText(),
			Behaviour: pqt.FunctionBehaviourImmutable,
		}, name)).
		AddColumn(pqt.NewColumn("age", pqt.TypeInteger())).
		AddColumn(pqt.NewColumn("created_at", pqt.TypeTimestampTZ(), pqt.WithNotNull())).
		AddColumn(name).
		AddColumn(description).
		AddUnique(name, description).
		AddCheck("name <> 'LOL'", name)

	pqt.NewSchema("constraints_test").
		AddTable(t1).
		AddTable(t2)

	g := &gogen.Generator{}
	g.Repository(t2) // Is here so output can be properly formatted
	g.RepositoryInsertQuery(t2)

	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) InsertQuery(e *T2Entity, read bool) (string, []interface{}, error) {
	insert := NewComposer(7)
	columns := bytes.NewBuffer(nil)
	buf := bytes.NewBufferString("INSERT INTO ")
	buf.WriteString(r.Table)

	if e.Age != nil {
		if columns.Len() > 0 {
			if _, err := columns.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := columns.WriteString(TableT2ColumnAge); err != nil {
			return "", nil, err
		}
		if insert.Dirty {
			if _, err := insert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := insert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		insert.Add(e.Age)
		insert.Dirty = true
	}

	if !e.CreatedAt.IsZero() {
		if columns.Len() > 0 {
			if _, err := columns.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := columns.WriteString(TableT2ColumnCreatedAt); err != nil {
			return "", nil, err
		}
		if insert.Dirty {
			if _, err := insert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := insert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		insert.Add(e.CreatedAt)
		insert.Dirty = true
	}

	if e.Description.Valid {
		if columns.Len() > 0 {
			if _, err := columns.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := columns.WriteString(TableT2ColumnDescription); err != nil {
			return "", nil, err
		}
		if insert.Dirty {
			if _, err := insert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := insert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		insert.Add(e.Description)
		insert.Dirty = true
	}

	if columns.Len() > 0 {
		if _, err := columns.WriteString(", "); err != nil {
			return "", nil, err
		}
	}
	if _, err := columns.WriteString(TableT2ColumnName); err != nil {
		return "", nil, err
	}
	if insert.Dirty {
		if _, err := insert.WriteString(", "); err != nil {
			return "", nil, err
		}
	}
	if err := insert.WritePlaceholder(); err != nil {
		return "", nil, err
	}
	insert.Add(e.Name)
	insert.Dirty = true

	if e.T1ID.Valid {
		if columns.Len() > 0 {
			if _, err := columns.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := columns.WriteString(TableT2ColumnT1ID); err != nil {
			return "", nil, err
		}
		if insert.Dirty {
			if _, err := insert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := insert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		insert.Add(e.T1ID)
		insert.Dirty = true
	}

	if columns.Len() > 0 {
		buf.WriteString(" (")
		buf.ReadFrom(columns)
		buf.WriteString(") VALUES (")
		buf.ReadFrom(insert)
		buf.WriteString(") ")
		if read {
			buf.WriteString("RETURNING ")
			if len(r.Columns) > 0 {
				buf.WriteString(strings.Join(r.Columns, ", "))
			} else {
				buf.WriteString("age, created_at, description, id, name, slugify() AS slug, t1_id")
			}
		}
	}
	return buf.String(), insert.Args(), nil
}`)
}
