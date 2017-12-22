package gogen_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/gogen"
)

func TestGenerator_RepositoryUpdateOneByPrimaryKey(t *testing.T) {
	t1 := pqt.NewTable("t1")
	t2 := pqt.NewTable("t2").
		AddRelationship(pqt.ManyToOne(t1)).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey()))

	g := &gogen.Generator{}
	g.Repository(t2)
	g.RepositoryUpdateOneByPrimaryKey(t2)
	assertOutput(t, g.Printer, `
type T2RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T2RepositoryBase) UpdateOneByID(ctx context.Context, pk int64, p *T2Patch) (*T2Entity, error) {
	query, args, err := r.UpdateOneByIDQuery(pk, p)
	if err != nil {
		return nil, err
	}
	var ent T2Entity
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
	if r.Log != nil {
		r.Log(err, TableT2, "update by primary key", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return &ent, nil
}`)
}

func TestGenerator_RepositoryUpdateOneByPrimaryKeyQuery(t *testing.T) {
	t0 := pqt.NewTable("t0")

	g := &gogen.Generator{}
	g.Repository(t0) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByPrimaryKeyQuery(t0)
	assertOutput(t, g.Printer, `
type T0RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)

	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey()))

	pqt.NewSchema("constraints_test").AddTable(t1)

	g = &gogen.Generator{}
	g.Repository(t1) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByPrimaryKeyQuery(t1)

	assertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) UpdateOneByIDQuery(pk int64, p *T1Patch) (string, []interface{}, error) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(r.Table)
	update := NewComposer(1)
	if !update.Dirty {
		return "", nil, errors.New("T1 update failure, nothing to update")
	}
	buf.WriteString(" SET ")
	buf.ReadFrom(update)
	buf.WriteString(" WHERE ")

	update.WriteString(TableT1ColumnID)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(pk)

	buf.ReadFrom(update)
	buf.WriteString(" RETURNING ")
	if len(r.Columns) > 0 {
		buf.WriteString(strings.Join(r.Columns, ", "))
	} else {
		buf.WriteString("id")
	}
	return buf.String(), update.Args(), nil
}`)
}

func TestGenerator_RepositoryUpdateOneByUniqueConstraintQuery(t *testing.T) {
	t0 := pqt.NewTable("t0")

	g := &gogen.Generator{}
	g.Repository(t0) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByUniqueConstraintQuery(t0)
	assertOutput(t, g.Printer, `
type T0RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)

	firstName := pqt.NewColumn("first_name", pqt.TypeText())
	lastName := pqt.NewColumn("last_name", pqt.TypeText())
	age := pqt.NewColumn("age", pqt.TypeIntegerBig())
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(firstName).
		AddColumn(lastName).
		AddColumn(age).
		AddUnique(firstName, lastName, age).
		AddUniqueIndex("AgeIsNotSet", "age IS NULL", firstName, lastName)

	pqt.NewSchema("constraints_test").AddTable(t1)

	g = &gogen.Generator{}
	g.Repository(t1) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByUniqueConstraintQuery(t1)

	assertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) UpdateOneByFirstNameAndLastNameAndAgeQuery(t1FirstName string, t1LastName string, t1Age int64, p *T1Patch) (string, []interface{}, error) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(r.Table)
	update := NewComposer(3)
	if p.Age.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnAge); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.Age)
		update.Dirty = true

	}

	if p.FirstName.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnFirstName); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.FirstName)
		update.Dirty = true

	}

	if p.LastName.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnLastName); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.LastName)
		update.Dirty = true

	}

	if !update.Dirty {
		return "", nil, errors.New("t1 update failure, nothing to update")
	}
	buf.WriteString(" SET ")
	buf.ReadFrom(update)
	buf.WriteString(" WHERE ")
	update.WriteString(TableT1ColumnFirstName)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(t1FirstName)
	update.WriteString(" AND ")
	update.WriteString(TableT1ColumnLastName)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(t1LastName)
	update.WriteString(" AND ")
	update.WriteString(TableT1ColumnAge)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(t1Age)
	buf.ReadFrom(update)
	buf.WriteString(" RETURNING ")
	if len(r.Columns) > 0 {
		buf.WriteString(strings.Join(r.Columns, ", "))
	} else {
		buf.WriteString("age, first_name, id, last_name")
	}
	return buf.String(), update.Args(), nil
}

func (r *T1RepositoryBase) UpdateOneByFirstNameAndLastNameWhereAgeIsNotSetQuery(t1FirstName string, t1LastName string, p *T1Patch) (string, []interface{}, error) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(r.Table)
	update := NewComposer(2)
	if p.Age.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnAge); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.Age)
		update.Dirty = true

	}

	if p.FirstName.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnFirstName); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.FirstName)
		update.Dirty = true

	}

	if p.LastName.Valid {
		if update.Dirty {
			if _, err := update.WriteString(", "); err != nil {
				return "", nil, err
			}
		}
		if _, err := update.WriteString(TableT1ColumnLastName); err != nil {
			return "", nil, err
		}
		if _, err := update.WriteString("="); err != nil {
			return "", nil, err
		}
		if err := update.WritePlaceholder(); err != nil {
			return "", nil, err
		}
		update.Add(p.LastName)
		update.Dirty = true

	}

	if !update.Dirty {
		return "", nil, errors.New("t1 update failure, nothing to update")
	}
	buf.WriteString(" SET ")
	buf.ReadFrom(update)
	buf.WriteString(" WHERE ")
	update.WriteString(TableT1ColumnFirstName)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(t1FirstName)
	update.WriteString(" AND ")
	update.WriteString(TableT1ColumnLastName)
	update.WriteString("=")
	update.WritePlaceholder()
	update.Add(t1LastName)
	buf.ReadFrom(update)
	buf.WriteString(" RETURNING ")
	if len(r.Columns) > 0 {
		buf.WriteString(strings.Join(r.Columns, ", "))
	} else {
		buf.WriteString("age, first_name, id, last_name WHERE age IS NULL")
	}
	return buf.String(), update.Args(), nil
}`)
}

func TestGenerator_RepositoryUpdateOneByUniqueConstraint(t *testing.T) {
	t0 := pqt.NewTable("t0")

	g := &gogen.Generator{}
	g.Repository(t0) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByUniqueConstraint(t0)
	assertOutput(t, g.Printer, `
type T0RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}`)

	firstName := pqt.NewColumn("first_name", pqt.TypeText())
	lastName := pqt.NewColumn("last_name", pqt.TypeText())
	age := pqt.NewColumn("age", pqt.TypeIntegerBig())
	t1 := pqt.NewTable("t1").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(firstName).
		AddColumn(lastName).
		AddColumn(age).
		AddUnique(firstName, lastName, age).
		AddUniqueIndex("AgeIsNotSet", "age IS NULL", firstName, lastName)

	pqt.NewSchema("constraints_test").AddTable(t1)

	g = &gogen.Generator{}
	g.Repository(t1) // Is here so output can be properly formatted
	g.RepositoryUpdateOneByUniqueConstraint(t1)

	assertOutput(t, g.Printer, `
type T1RepositoryBase struct {
	Table   string
	Columns []string
	DB      *sql.DB
	Log     LogFunc
}

func (r *T1RepositoryBase) UpdateOneByFirstNameAndLastNameAndAge(ctx context.Context, t1FirstName string, t1LastName string, t1Age int64, p *T1Patch) (*T1Entity, error) {
	query, args, err := r.UpdateOneByFirstNameAndLastNameAndAgeQuery(t1FirstName, t1LastName, t1Age, p)
	if err != nil {
		return nil, err
	}
	var ent T1Entity
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
	if r.Log != nil {
		r.Log(err, TableT1, "update one by unique", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

func (r *T1RepositoryBase) UpdateOneByFirstNameAndLastNameWhereAgeIsNotSet(ctx context.Context, t1FirstName string, t1LastName string, p *T1Patch) (*T1Entity, error) {
	query, args, err := r.UpdateOneByFirstNameAndLastNameWhereAgeIsNotSetQuery(t1FirstName, t1LastName, p)
	if err != nil {
		return nil, err
	}
	var ent T1Entity
	props, err := ent.Props(r.Columns...)
	if err != nil {
		return nil, err
	}
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
	if r.Log != nil {
		r.Log(err, TableT1, "update one by unique", query, args...)
	}
	if err != nil {
		return nil, err
	}
	return &ent, nil
}`)
}
