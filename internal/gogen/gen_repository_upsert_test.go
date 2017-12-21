package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_RepositoryUpsertQuery(t *testing.T) {
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

	g := &gogen.Generator{Version: 9.4}
	g.Repository(t2) // Is here so output can be properly formatted
	g.RepositoryUpsertQuery(t2)

	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)
	g = &gogen.Generator{Version: 9.5}
	g.Repository(t2) // Is here so output can be properly formatted
	g.RepositoryUpsertQuery(t2)

	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) UpsertQuery(e *T2Entity, p *T2Patch, inf ...string) (string, []interface{}, error) {
	upsert := NewComposer(14)
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
		if upsert.Dirty {
			if _, err := upsert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := upsert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		upsert.Add(e.Age)
		upsert.Dirty = true
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
		if upsert.Dirty {
			if _, err := upsert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := upsert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		upsert.Add(e.CreatedAt)
		upsert.Dirty = true
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
		if upsert.Dirty {
			if _, err := upsert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := upsert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		upsert.Add(e.Description)
		upsert.Dirty = true
	}

	if columns.Len() > 0 {
		if _, err := columns.WriteString(", "); err != nil {
			return "", nil, err
		}
	}
	if _, err := columns.WriteString(TableT2ColumnName); err != nil {
		return "", nil, err
	}
	if upsert.Dirty {
		if _, err := upsert.WriteString(", "); err != nil {
			return "", nil, err
		}
	}
	if err := upsert.WritePlaceholder(); err != nil {
		return "", nil, err
	}
	upsert.Add(e.Name)
	upsert.Dirty = true

	if e.T1ID.Valid {
		if columns.Len() > 0 {
			if _, err := columns.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := columns.WriteString(TableT2ColumnT1ID); err != nil {
			return "", nil, err
		}
		if upsert.Dirty {
			if _, err := upsert.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if err := upsert.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		upsert.Add(e.T1ID)
		upsert.Dirty = true
	}

	if upsert.Dirty {
		buf.WriteString(" (")
		buf.ReadFrom(columns)
		buf.WriteString(") VALUES (")
		buf.ReadFrom(upsert)
		buf.WriteString(")")
	}
	buf.WriteString(" ON CONFLICT ")
	if len(inf) > 0 {
		upsert.Dirty = false
		if p.Age != nil {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnAge); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.Age)
			upsert.Dirty = true

		}

		if p.CreatedAt.Valid {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnCreatedAt); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.CreatedAt)
			upsert.Dirty = true

		}

		if p.Description.Valid {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnDescription); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.Description)
			upsert.Dirty = true

		}

		if p.ID.Valid {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnID); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.ID)
			upsert.Dirty = true

		}

		if p.Name.Valid {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnName); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.Name)
			upsert.Dirty = true

		}

		if p.T1ID.Valid {
			if upsert.Dirty {
				if _, err := upsert.WriteString(", "); err != nil {
					return "", nil, err
				}
			}
			if _, err := upsert.WriteString(TableT2ColumnT1ID); err != nil {
				return "", nil, err
			}
			if _, err := upsert.WriteString("="); err != nil {
				return "", nil, err
			}
			if err := upsert.WritePlaceholder(); err != nil {
				return "", nil, err
			}
			upsert.Add(p.T1ID)
			upsert.Dirty = true

		}

	}

	if len(inf) > 0 && upsert.Dirty {
		buf.WriteString("(")
		for j, i := range inf {
			if j != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(i)
		}
		buf.WriteString(")")
		buf.WriteString(" DO UPDATE SET ")
		buf.ReadFrom(upsert)
	} else {
		buf.WriteString(" DO NOTHING ")
	}
	if upsert.Dirty {
		buf.WriteString(" RETURNING ")
		if len(r.Columns) > 0 {
			buf.WriteString(strings.Join(r.Columns, ", "))
		} else {
			buf.WriteString("age, created_at, description, id, name, slugify() AS slug, t1_id")
		}
	}
	return buf.String(), upsert.Args(), nil
}`)
}
