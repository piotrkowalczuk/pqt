package pqtgogen_test

import (
	"bytes"
	"go/format"
	"testing"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/internal/testutil"
	"github.com/piotrkowalczuk/pqt/pqtgo/pqtgogen"
)

func TestGenerator(t *testing.T) {
	cases := map[string]struct {
		components pqtgogen.Component
		schema     func() *pqt.Schema
		expected   string
	}{
		"simple-helpers": {
			components: pqtgogen.ComponentAll,
			schema: func() *pqt.Schema {
				userID := pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())
				user := pqt.NewTable("user", pqt.WithTableIfNotExists()).
					AddColumn(userID).
					AddColumn(pqt.NewColumn("name", pqt.TypeText(),
						pqt.WithNotNull(),
						pqt.WithUnique(),

						pqt.WithDefault("'empty'"),
						pqt.WithCheck("name <> 'something'"),
					))
				post := pqt.NewTable("post").
					AddColumn(pqt.NewColumn("body", pqt.TypeBytea()))
				comment := pqt.NewTable("comment").
					AddColumn(pqt.NewColumn("user_id", pqt.TypeIntegerBig(), pqt.WithReference(userID))).
					AddRelationship(pqt.ManyToOne(post, pqt.WithOwnerName("komentarz"), pqt.WithInversedName("wpis")))

				return pqt.NewSchema("example").AddTable(user).AddTable(comment)
			},
			expected: expectedSimple,
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			g := pqtgogen.Generator{
				Version:    9.5,
				Pkg:        "example",
				Components: c.components,
			}
			s := c.schema()
			buf, err := g.Generate(s)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			got := normalize(t, buf)
			expected := normalize(t, []byte(c.expected))
			testutil.AssertGoCode(t, expected, got)

			into := bytes.NewBuffer(nil)
			err = g.GenerateTo(s, into)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			testutil.AssertGoCode(t, expected, into.String())
		})
	}
}

func normalize(t *testing.T, in []byte) string {
	out, err := format.Source(in)
	if err != nil {
		t.Fatalf("formating failure: %s", err.Error())
	}
	return string(out)
}

var expectedSimple = `package example

    		import (
    			"github.com/m4rw3r/uuid"
    		)

    		// LogFunc represents function that can be passed into repository to log query result.
    		type LogFunc func(err error, ent, fnc, sql string, args ...interface{})

    		// Rows ...
    		type Rows interface {
    			io.Closer
    			ColumnTypes() ([]*sql.ColumnType, error)
    			Columns() ([]string, error)
    			Err() error
    			Next() bool
    			NextResultSet() bool
    			Scan(dst ...interface{}) error
    		}

    		func joinClause(comp *Composer, jt JoinType, on string) (ok bool, err error) {
    			if jt != JoinDoNot {
    				switch jt {
    				case JoinInner:
    					if _, err = comp.WriteString(" INNER JOIN "); err != nil {
    						return
    					}
    				case JoinLeft:
    					if _, err = comp.WriteString(" LEFT JOIN "); err != nil {
    						return
    					}
    				case JoinRight:
    					if _, err = comp.WriteString(" RIGHT JOIN "); err != nil {
    						return
    					}
    				case JoinCross:
    					if _, err = comp.WriteString(" CROSS JOIN "); err != nil {
    						return
    					}
    				default:
    					return
    				}
    				if _, err = comp.WriteString(on); err != nil {
    					return
    				}
    				comp.Dirty = true
    				ok = true
    				return
    			}
    			return
    		}

    		const (
    			TableUserConstraintPrimaryKey = "example.user_id_pkey"
    			TableUserConstraintNameUnique = "example.user_name_key"
    			TableUserConstraintNameCheck = "example.user_name_check"
    		)

    		const (
    			TableUser                     = "example.user"
    			TableUserColumnID             = "id"
    			TableUserColumnName           = "name"
    		)

    		var TableUserColumns = []string{
				TableUserColumnID,
				TableUserColumnName,
			}

    		// UserEntity ...
    		type UserEntity struct {
    			// ID ...
    			ID int64
    			// Name ...
    			Name string
    		}

    		func (e *UserEntity) Prop(cn string) (interface{}, bool) {
    			switch cn {

    			case TableUserColumnID:
    				return &e.ID, true
    			case TableUserColumnName:
    				return &e.Name, true
    			default:
    				return nil, false
    			}
    		}

    		func (e *UserEntity) Props(cns ...string) ([]interface{}, error) {
    			if len(cns) == 0 {
    				cns = TableUserColumns
    			}
    			res := make([]interface{}, 0, len(cns))
    			for _, cn := range cns {
    				if prop, ok := e.Prop(cn); ok {
    					res = append(res, prop)
    				} else {
    					return nil, fmt.Errorf("unexpected column provided: %s", cn)
    				}
    			}
    			return res, nil
    		}

    		// ScanUserRows helps to scan rows straight to the slice of entities.
    		func ScanUserRows(rows Rows) (entities []*UserEntity, err error) {
    			for rows.Next() {
    				var ent UserEntity
    				err = rows.Scan(
    					&ent.ID,
    					&ent.Name,
    				)
    				if err != nil {
    					return
    				}

    				entities = append(entities, &ent)
    			}
    			if err = rows.Err(); err != nil {
    				return
    			}

    			return
    		}

    		// UserIterator is not thread safe.
    		type UserIterator struct {
    			rows Rows
    			cols []string
    			expr *UserFindExpr
    		}

    		func (i *UserIterator) Next() bool {
    			return i.rows.Next()
    		}

    		func (i *UserIterator) Close() error {
    			return i.rows.Close()
    		}

    		func (i *UserIterator) Err() error {
    			return i.rows.Err()
    		}

    		// Columns is wrapper around sql.Rows.Columns method, that also cache output inside iterator.
    		func (i *UserIterator) Columns() ([]string, error) {
    			if i.cols == nil {
    				cols, err := i.rows.Columns()
    				if err != nil {
    					return nil, err
    				}
    				i.cols = cols
    			}
    			return i.cols, nil
    		}

    		// Ent is wrapper around User method that makes iterator more generic.
    		func (i *UserIterator) Ent() (interface{}, error) {
    			return i.User()
    		}

    		func (i *UserIterator) User() (*UserEntity, error) {
    			var ent UserEntity
    			cols, err := i.Columns()
    			if err != nil {
    				return nil, err
    			}

    			props, err := ent.Props(cols...)
    			if err != nil {
    				return nil, err
    			}
    			if err := i.rows.Scan(props...); err != nil {
    				return nil, err
    			}
    			return &ent, nil
    		}

    		type UserCriteria struct {
    			ID   sql.NullInt64
    			Name sql.NullString
				operator string
				child, sibling, parent *UserCriteria
    		}

			func UserOperand(operator string, operands ...*UserCriteria) *UserCriteria {
				if len(operands) == 0 {
					return &UserCriteria{operator: operator}
				}

				parent := &UserCriteria{
					operator: operator,
					child:    operands[0],
				}

				for i := 0; i < len(operands); i++ {
					if i < len(operands)-1 {
						operands[i].sibling = operands[i+1]
					}
					operands[i].parent = parent
				}

				return parent
			}

			func UserOr(operands ...*UserCriteria) *UserCriteria {
				return UserOperand("OR", operands...)
			}

			func UserAnd(operands ...*UserCriteria) *UserCriteria {
				return UserOperand("AND", operands...)
			}

    		type UserFindExpr struct {
    			Where         *UserCriteria
    			Offset, Limit int64
    			Columns       []string
    			OrderBy       []RowOrder
    		}

    		type UserJoin struct {
    			On, Where *UserCriteria
    			Fetch     bool
    			Kind      JoinType
    		}

    		type UserCountExpr struct {
    			Where *UserCriteria
    		}

    		type UserPatch struct {
    			Name sql.NullString
    		}

    		type UserRepositoryBase struct {
    			Table   string
    			Columns []string
    			DB      *sql.DB
    			Log     LogFunc
    		}

    		func (r *UserRepositoryBase) InsertQuery(e *UserEntity, read bool) (string, []interface{}, error) {
    			insert := NewComposer(2)
    			columns := bytes.NewBuffer(nil)
    			buf := bytes.NewBufferString("INSERT INTO ")
    			buf.WriteString(r.Table)

    			if columns.Len() > 0 {
    				if _, err := columns.WriteString(", "); err != nil {
    					return "", nil, err
    				}
    			}
    			if _, err := columns.WriteString(TableUserColumnName); err != nil {
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
    						buf.WriteString("id, name")
    					}
    				}
    			}
    			return buf.String(), insert.Args(), nil
    		}

    		func (r *UserRepositoryBase) Insert(ctx context.Context, e *UserEntity) (*UserEntity, error) {
    			query, args, err := r.InsertQuery(e, true)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(
					&e.ID,
    				&e.Name,
    			)
    			if r.Log != nil {
    				r.Log(err, TableUser, "insert", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return e, nil
    		}

			func UserCriteriaWhereClause(comp *Composer, c *UserCriteria, id int) error {
				if c.child == nil {
					return _UserCriteriaWhereClause(comp, c, id)
				}
				node := c
				sibling := false
				for {
					if !sibling {
						if node.child != nil {
							if node.parent != nil {
								comp.WriteString("(")
							}
							node = node.child
							continue
						} else {
							comp.Dirty = false
							comp.WriteString("(")
							if err := _UserCriteriaWhereClause(comp, node, id); err != nil {
								return err
							}
							comp.WriteString(")")
						}
					}
					if node.sibling != nil {
						sibling = false
						comp.WriteString(" ")
						comp.WriteString(node.parent.operator)
						comp.WriteString(" ")
						node = node.sibling
						continue
					}
					if node.parent != nil {
						sibling = true
						if node.parent.parent != nil {
							comp.WriteString(")")
						}
						node = node.parent
						continue
					}
			
					break
				}
				return nil
			}

    		func _UserCriteriaWhereClause(comp *Composer, c *UserCriteria, id int) error {
    			if c.ID.Valid {
    				if comp.Dirty {
    					comp.WriteString(" AND ")
    				}
    				if err := comp.WriteAlias(id); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString(TableUserColumnID); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString("="); err != nil {
    					return err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return err
    				}
    				comp.Add(c.ID)
    				comp.Dirty = true
    			}
    			if c.Name.Valid {
    				if comp.Dirty {
    					comp.WriteString(" AND ")
    				}
    				if err := comp.WriteAlias(id); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString(TableUserColumnName); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString("="); err != nil {
    					return err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return err
    				}
    				comp.Add(c.Name)
    				comp.Dirty = true
    			}
    			return nil
    		}

    		func (r *UserRepositoryBase) FindQuery(fe *UserFindExpr) (string, []interface{}, error) {
    			comp := NewComposer(2)
    			buf := bytes.NewBufferString("SELECT ")
    			if len(fe.Columns) == 0 {
    				buf.WriteString("t0.id, t0.name")
    			} else {
    				buf.WriteString(strings.Join(fe.Columns, ", "))
    			}
    			buf.WriteString(" FROM ")
    			buf.WriteString(r.Table)
    			buf.WriteString(" AS t0")
    			if comp.Dirty {
    				buf.ReadFrom(comp)
    				comp.Dirty = false
    			}
    			if fe.Where != nil {
    				if err := UserCriteriaWhereClause(comp, fe.Where, 0); err != nil {
    					return "", nil, err
    				}
    			}
    			if comp.Dirty {
    				if _, err := buf.WriteString(" WHERE "); err != nil {
    					return "", nil, err
    				}
    				buf.ReadFrom(comp)
    			}

    			if len(fe.OrderBy) > 0 {
    				i := 0
    				for _, order := range fe.OrderBy {
    					for _, columnName := range TableUserColumns {
    						if order.Name == columnName {
    							if i == 0 {
    								comp.WriteString(" ORDER BY ")
    							}
    							if i > 0 {
    								if _, err := comp.WriteString(", "); err != nil {
    									return "", nil, err
    								}
    							}
    							if _, err := comp.WriteString(order.Name); err != nil {
    								return "", nil, err
    							}
    							if order.Descending {
    								if _, err := comp.WriteString(" DESC"); err != nil {
    									return "", nil, err
    								}
    							}
    							i++
    							break
    						}
    					}
    				}
    			}
    			if fe.Offset > 0 {
    				if _, err := comp.WriteString(" OFFSET "); err != nil {
    					return "", nil, err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				if _, err := comp.WriteString(" "); err != nil {
    					return "", nil, err
    				}
    				comp.Add(fe.Offset)
    			}
    			if fe.Limit > 0 {
    				if _, err := comp.WriteString(" LIMIT "); err != nil {
    					return "", nil, err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				if _, err := comp.WriteString(" "); err != nil {
    					return "", nil, err
    				}
    				comp.Add(fe.Limit)
    			}

    			buf.ReadFrom(comp)

    			return buf.String(), comp.Args(), nil
    		}

    		func (r *UserRepositoryBase) Find(ctx context.Context, fe *UserFindExpr) ([]*UserEntity, error) {
    			query, args, err := r.FindQuery(fe)
    			if err != nil {
    				return nil, err
    			}
    			rows, err := r.DB.QueryContext(ctx, query, args...)
    			if r.Log != nil {
    				r.Log(err, TableUser, "find", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			defer rows.Close()
    			var (
					entities []*UserEntity
    				props []interface{}
				)
    			for rows.Next() {
    				var ent UserEntity
    				if props, err = ent.Props(); err != nil {
    					return nil, err
    				}
    				err = rows.Scan(props...)
    				if err != nil {
    					return nil, err
    				}

    				entities = append(entities, &ent)
    			}
    			err = rows.Err()
    			if r.Log != nil {
    				r.Log(err, TableUser, "find", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return entities, nil
    		}

    		func (r *UserRepositoryBase) FindIter(ctx context.Context, fe *UserFindExpr) (*UserIterator, error) {
    			query, args, err := r.FindQuery(fe)
    			if err != nil {
    				return nil, err
    			}
    			rows, err := r.DB.QueryContext(ctx, query, args...)
    			if r.Log != nil {
    				r.Log(err, TableUser, "find iter", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return &UserIterator{
    				rows: rows,
    				expr: fe,
    				cols: fe.Columns,
    			}, nil
    		}

    		func (r *UserRepositoryBase) FindOneByID(ctx context.Context, pk int64) (*UserEntity, error) {
    			find := NewComposer(2)
    			find.WriteString("SELECT ")
    			if len(r.Columns) == 0 {
    				find.WriteString("id, name")
    			} else {
    				find.WriteString(strings.Join(r.Columns, ", "))
    			}
    			find.WriteString(" FROM ")
    			find.WriteString(TableUser)
    			find.WriteString(" WHERE ")
    			find.WriteString(TableUserColumnID)
    			find.WriteString("=")
    			find.WritePlaceholder()
    			find.Add(pk)
    			var (
    				ent UserEntity
    			)
    			props, err := ent.Props(r.Columns...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)
    			if r.Log != nil {
    				r.Log(err, TableUser, "find by primary key", find.String(), find.Args()...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return &ent, nil
    		}

    		func (r *UserRepositoryBase) FindOneByName(ctx context.Context, userName string) (*UserEntity, error) {
    			find := NewComposer(2)
    			find.WriteString("SELECT ")
    			if len(r.Columns) == 0 {
    				find.WriteString("id, name")
    			} else {
    				find.WriteString(strings.Join(r.Columns, ", "))
    			}
    			find.WriteString(" FROM ")
    			find.WriteString(TableUser)
    			find.WriteString(" WHERE ")
    			find.WriteString(TableUserColumnName)
    			find.WriteString("=")
    			find.WritePlaceholder()
    			find.Add(userName)

    			var (
    				ent UserEntity
    			)
    			props, err := ent.Props(r.Columns...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, find.String(), find.Args()...).Scan(props...)
    			if err != nil {
    				return nil, err
    			}

    			return &ent, nil
    		}

    		func (r *UserRepositoryBase) UpdateOneByIDQuery(pk int64, p *UserPatch) (string, []interface{}, error) {
    			buf := bytes.NewBufferString("UPDATE ")
    			buf.WriteString(r.Table)
    			update := NewComposer(2)
    			if p.Name.Valid {
    				if update.Dirty {
    					if _, err := update.WriteString(", "); err != nil {
    						return "", nil, err
    					}
    				}
    				if _, err := update.WriteString(TableUserColumnName); err != nil {
    					return "", nil, err
    				}
    				if _, err := update.WriteString("="); err != nil {
    					return "", nil, err
    				}
    				if err := update.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				update.Add(p.Name)
    				update.Dirty = true

    			}
    			if !update.Dirty {
    				return "", nil, errors.New("User update failure, nothing to update")
    			}
    			buf.WriteString(" SET ")
    			buf.ReadFrom(update)
    			buf.WriteString(" WHERE ")

    			update.WriteString(TableUserColumnID)
    			update.WriteString("=")
    			update.WritePlaceholder()
    			update.Add(pk)

    			buf.ReadFrom(update)
    			buf.WriteString(" RETURNING ")
    			if len(r.Columns) > 0 {
    				buf.WriteString(strings.Join(r.Columns, ", "))
    			} else {
    				buf.WriteString("id, name")
    			}
    			return buf.String(), update.Args(), nil
    		}

    		func (r *UserRepositoryBase) UpdateOneByID(ctx context.Context, pk int64, p *UserPatch) (*UserEntity, error) {
    			query, args, err := r.UpdateOneByIDQuery(pk, p)
    			if err != nil {
    				return nil, err
    			}
    			var ent UserEntity
    			props, err := ent.Props(r.Columns...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
    			if r.Log != nil {
    				r.Log(err, TableUser, "update by primary key", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return &ent, nil
    		}

		func (r *UserRepositoryBase) FetchAndUpdateOneByID(ctx context.Context, pk int64, p *UserPatch) (before, after *UserEntity, err error) {
			find := NewComposer(2)
			find.WriteString("SELECT ")
			if len(r.Columns) == 0 {
				find.WriteString("id, name")
			} else {
				find.WriteString(strings.Join(r.Columns, ", "))
			}
			find.WriteString(" FROM ")
			find.WriteString(TableUser)
			find.WriteString(" WHERE ")
			find.WriteString(TableUserColumnID)
			find.WriteString("=")
			find.WritePlaceholder()
			find.Add(pk)
			find.WriteString(" FOR UPDATE")
			query, args, err := r.UpdateOneByIDQuery(pk, p)
			if err != nil {
				return
			}
			var (
				oldEnt, newEnt UserEntity
			)
			oldProps, err := oldEnt.Props(r.Columns...)
			if err != nil {
				return
			}
			newProps, err := newEnt.Props(r.Columns...)
			if err != nil {
				return
			}
			tx, err := r.DB.Begin()
			if err != nil {
				return
			}
			err = tx.QueryRowContext(ctx, find.String(), find.Args()...).Scan(oldProps...)
			if r.Log != nil {
				r.Log(err, TableUser, "find by primary key", find.String(), find.Args()...)
			}
			if err != nil {
				tx.Rollback()
				return
			}
			err = tx.QueryRowContext(ctx, query, args...).Scan(newProps...)
			if r.Log != nil {
				r.Log(err, TableUser, "update by primary key", query, args...)
			}
			if err != nil {
				tx.Rollback()
				return
			}
			err = tx.Commit()
			if err != nil {
				return
			}
			return &oldEnt, &newEnt, nil
		}

    		func (r *UserRepositoryBase) UpdateOneByNameQuery(userName string, p *UserPatch) (string, []interface{}, error) {
    			buf := bytes.NewBufferString("UPDATE ")
    			buf.WriteString(r.Table)
    			update := NewComposer(1)
    			if p.Name.Valid {
    				if update.Dirty {
    					if _, err := update.WriteString(", "); err != nil {
    						return "", nil, err
    					}
    				}
    				if _, err := update.WriteString(TableUserColumnName); err != nil {
    					return "", nil, err
    				}
    				if _, err := update.WriteString("="); err != nil {
    					return "", nil, err
    				}
    				if err := update.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				update.Add(p.Name)
    				update.Dirty = true

    			}
    			if !update.Dirty {
    				return "", nil, errors.New("user update failure, nothing to update")
    			}
    			buf.WriteString(" SET ")
    			buf.ReadFrom(update)
    			buf.WriteString(" WHERE ")
    			update.WriteString(TableUserColumnName)
    			update.WriteString("=")
    			update.WritePlaceholder()
    			update.Add(userName)
    			buf.ReadFrom(update)
    			buf.WriteString(" RETURNING ")
    			if len(r.Columns) > 0 {
    				buf.WriteString(strings.Join(r.Columns, ", "))
    			} else {
    				buf.WriteString("id, name")
    			}
    			return buf.String(), update.Args(), nil
    		}

    		func (r *UserRepositoryBase) UpdateOneByName(ctx context.Context, userName string, p *UserPatch) (*UserEntity, error) {
    			query, args, err := r.UpdateOneByNameQuery(userName, p)
    			if err != nil {
    				return nil, err
    			}
    			var ent UserEntity
    			props, err := ent.Props(r.Columns...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(props...)
    			if r.Log != nil {
    				r.Log(err, TableUser, "update one by unique", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return &ent, nil
    		}

    		func (r *UserRepositoryBase) UpsertQuery(e *UserEntity, p *UserPatch, inf ...string) (string, []interface{}, error) {
    			upsert := NewComposer(4)
    			columns := bytes.NewBuffer(nil)
    			buf := bytes.NewBufferString("INSERT INTO ")
    			buf.WriteString(r.Table)

    			if columns.Len() > 0 {
    				if _, err := columns.WriteString(", "); err != nil {
    					return "", nil, err
    				}
    			}
    			if _, err := columns.WriteString(TableUserColumnName); err != nil {
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
    				if p.Name.Valid {
    					if upsert.Dirty {
    						if _, err := upsert.WriteString(", "); err != nil {
    							return "", nil, err
    						}
    					}
    					if _, err := upsert.WriteString(TableUserColumnName); err != nil {
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
    					buf.WriteString("id, name")
    				}
    			}
    			return buf.String(), upsert.Args(), nil
    		}

    		func (r *UserRepositoryBase) Upsert(ctx context.Context, e *UserEntity, p *UserPatch, inf ...string) (*UserEntity, error) {
    			query, args, err := r.UpsertQuery(e, p, inf...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(
					&e.ID,
    				&e.Name,
    			)
    			if r.Log != nil {
    				r.Log(err, TableUser, "upsert", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return e, nil
    		}

    		func (r *UserRepositoryBase) Count(ctx context.Context, c *UserCountExpr) (int64, error) {
    			query, args, err := r.FindQuery(&UserFindExpr{
    				Where:   c.Where,
    				Columns: []string{"COUNT(*)"},
    			})
    			if err != nil {
    				return 0, err
    			}
    			var count int64
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(&count)
    			if r.Log != nil {
    				r.Log(err, TableUser, "count", query, args...)
    			}
    			if err != nil {
    				return 0, err
    			}
    			return count, nil
    		}

    		func (r *UserRepositoryBase) DeleteOneByID(ctx context.Context, pk int64) (int64, error) {
    			find := NewComposer(2)
    			find.WriteString("DELETE FROM ")
    			find.WriteString(TableUser)
    			find.WriteString(" WHERE ")
    			find.WriteString(TableUserColumnID)
    			find.WriteString("=")
    			find.WritePlaceholder()
    			find.Add(pk)
    			res, err := r.DB.ExecContext(ctx, find.String(), find.Args()...)
    			if err != nil {
    				return 0, err
    			}

    			return res.RowsAffected()
    		}

    		const (
    			TableCommentConstraintUserIDForeignKey = "example.comment_user_id_fkey"
    		)

    		const (
    			TableComment                           = "example.comment"
    			TableCommentColumnUserID               = "user_id"
    		)

    		var TableCommentColumns = []string{
				TableCommentColumnUserID,
			}

    		// CommentEntity ...
    		type CommentEntity struct {
    			// UserID ...
    			UserID sql.NullInt64
    			// User ...
    			User *UserEntity
    			// Wpis ...
    			Wpis *PostEntity
    		}

    		func (e *CommentEntity) Prop(cn string) (interface{}, bool) {
    			switch cn {

    			case TableCommentColumnUserID:
    				return &e.UserID, true
    			default:
    				return nil, false
    			}
    		}

    		func (e *CommentEntity) Props(cns ...string) ([]interface{}, error) {
    			if len(cns) == 0 {
    				cns = TableCommentColumns
    			}
    			res := make([]interface{}, 0, len(cns))
    			for _, cn := range cns {
    				if prop, ok := e.Prop(cn); ok {
    					res = append(res, prop)
    				} else {
    					return nil, fmt.Errorf("unexpected column provided: %s", cn)
    				}
    			}
    			return res, nil
    		}

    		// ScanCommentRows helps to scan rows straight to the slice of entities.
    		func ScanCommentRows(rows Rows) (entities []*CommentEntity, err error) {
    			for rows.Next() {
    				var ent CommentEntity
    				err = rows.Scan(
    					&ent.UserID,
    				)
    				if err != nil {
    					return
    				}

    				entities = append(entities, &ent)
    			}
    			if err = rows.Err(); err != nil {
    				return
    			}

    			return
    		}

    		// CommentIterator is not thread safe.
    		type CommentIterator struct {
    			rows Rows
    			cols []string
    			expr *CommentFindExpr
    		}

    		func (i *CommentIterator) Next() bool {
    			return i.rows.Next()
    		}

    		func (i *CommentIterator) Close() error {
    			return i.rows.Close()
    		}

    		func (i *CommentIterator) Err() error {
    			return i.rows.Err()
    		}

    		// Columns is wrapper around sql.Rows.Columns method, that also cache output inside iterator.
    		func (i *CommentIterator) Columns() ([]string, error) {
    			if i.cols == nil {
    				cols, err := i.rows.Columns()
    				if err != nil {
    					return nil, err
    				}
    				i.cols = cols
    			}
    			return i.cols, nil
    		}

    		// Ent is wrapper around Comment method that makes iterator more generic.
    		func (i *CommentIterator) Ent() (interface{}, error) {
    			return i.Comment()
    		}

    		func (i *CommentIterator) Comment() (*CommentEntity, error) {
    			var ent CommentEntity
    			cols, err := i.Columns()
    			if err != nil {
    				return nil, err
    			}

    			props, err := ent.Props(cols...)
    			if err != nil {
    				return nil, err
    			}
    			var prop []interface{}
    			if i.expr.JoinUser != nil && i.expr.JoinUser.Kind.Actionable() && i.expr.JoinUser.Fetch {
    				ent.User = &UserEntity{}
    				if prop, err = ent.User.Props(); err != nil {
    					return nil, err
    				}
    				props = append(props, prop...)
    			}
    			if i.expr.JoinWpis != nil && i.expr.JoinWpis.Kind.Actionable() && i.expr.JoinWpis.Fetch {
    				ent.Wpis = &PostEntity{}
    				if prop, err = ent.Wpis.Props(); err != nil {
    					return nil, err
    				}
    				props = append(props, prop...)
    			}
    			if err := i.rows.Scan(props...); err != nil {
    				return nil, err
    			}
    			return &ent, nil
    		}

    		type CommentCriteria struct {
    			UserID sql.NullInt64
				operator string
				child, sibling, parent *CommentCriteria
    		}

			func CommentOperand(operator string, operands ...*CommentCriteria) *CommentCriteria {
				if len(operands) == 0 {
					return &CommentCriteria{operator: operator}
				}

				parent := &CommentCriteria{
					operator: operator,
					child:    operands[0],
				}

				for i := 0; i < len(operands); i++ {
					if i < len(operands)-1 {
						operands[i].sibling = operands[i+1]
					}
					operands[i].parent = parent
				}

				return parent
			}

			func CommentOr(operands ...*CommentCriteria) *CommentCriteria {
				return CommentOperand("OR", operands...)
			}

			func CommentAnd(operands ...*CommentCriteria) *CommentCriteria {
				return CommentOperand("AND", operands...)
			}

    		type CommentFindExpr struct {
    			Where         *CommentCriteria
    			Offset, Limit int64
    			Columns       []string
    			OrderBy       []RowOrder
    			JoinUser      *UserJoin
    			JoinWpis      *PostJoin
    		}

    		type CommentJoin struct {
    			On, Where *CommentCriteria
    			Fetch     bool
    			Kind      JoinType
    			JoinUser  *UserJoin
    			JoinWpis  *PostJoin
    		}

    		type CommentCountExpr struct {
    			Where    *CommentCriteria
    			JoinUser *UserJoin
    			JoinWpis *PostJoin
    		}

    		type CommentPatch struct {
    			UserID sql.NullInt64
    		}

    		type CommentRepositoryBase struct {
    			Table   string
    			Columns []string
    			DB      *sql.DB
    			Log     LogFunc
    		}

    		func (r *CommentRepositoryBase) InsertQuery(e *CommentEntity, read bool) (string, []interface{}, error) {
    			insert := NewComposer(1)
    			columns := bytes.NewBuffer(nil)
    			buf := bytes.NewBufferString("INSERT INTO ")
    			buf.WriteString(r.Table)

    			if e.UserID.Valid {
    				if columns.Len() > 0 {
    					if _, err := columns.WriteString(", "); err != nil {
    						return "", nil, err
    					}
    				}
    				if _, err := columns.WriteString(TableCommentColumnUserID); err != nil {
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
    				insert.Add(e.UserID)
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
    						buf.WriteString("user_id")
    					}
    				}
    			}
    			return buf.String(), insert.Args(), nil
    		}

    		func (r *CommentRepositoryBase) Insert(ctx context.Context, e *CommentEntity) (*CommentEntity, error) {
    			query, args, err := r.InsertQuery(e, true)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(
					&e.UserID,
				)
    			if r.Log != nil {
    				r.Log(err, TableComment, "insert", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return e, nil
    		}

			func CommentCriteriaWhereClause(comp *Composer, c *CommentCriteria, id int) error {
				if c.child == nil {
					return _CommentCriteriaWhereClause(comp, c, id)
				}
				node := c
				sibling := false
				for {
					if !sibling {
						if node.child != nil {
							if node.parent != nil {
								comp.WriteString("(")
							}
							node = node.child
							continue
						} else {
							comp.Dirty = false
							comp.WriteString("(")
							if err := _CommentCriteriaWhereClause(comp, node, id); err != nil {
								return err
							}
							comp.WriteString(")")
						}
					}
					if node.sibling != nil {
						sibling = false
						comp.WriteString(" ")
						comp.WriteString(node.parent.operator)
						comp.WriteString(" ")
						node = node.sibling
						continue
					}
					if node.parent != nil {
						sibling = true
						if node.parent.parent != nil {
							comp.WriteString(")")
						}
						node = node.parent
						continue
					}

					break
				}
				return nil
			}

    		func _CommentCriteriaWhereClause(comp *Composer, c *CommentCriteria, id int) error {
    			if c.UserID.Valid {
    				if comp.Dirty {
    					comp.WriteString(" AND ")
    				}
    				if err := comp.WriteAlias(id); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString(TableCommentColumnUserID); err != nil {
    					return err
    				}
    				if _, err := comp.WriteString("="); err != nil {
    					return err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return err
    				}
    				comp.Add(c.UserID)
    				comp.Dirty = true
    			}
    			return nil
    		}

    		func (r *CommentRepositoryBase) FindQuery(fe *CommentFindExpr) (string, []interface{}, error) {
    			comp := NewComposer(1)
    			buf := bytes.NewBufferString("SELECT ")
    			if len(fe.Columns) == 0 {
    				buf.WriteString("t0.user_id")
    			} else {
    				buf.WriteString(strings.Join(fe.Columns, ", "))
    			}
    			if fe.JoinUser != nil && fe.JoinUser.Kind.Actionable() && fe.JoinUser.Fetch {
    				buf.WriteString(", t1.id, t1.name")
    			}
    			if fe.JoinWpis != nil && fe.JoinWpis.Kind.Actionable() && fe.JoinWpis.Fetch {
    				buf.WriteString(", t2.body")
    			}
    			buf.WriteString(" FROM ")
    			buf.WriteString(r.Table)
    			buf.WriteString(" AS t0")
    			if fe.JoinUser != nil && fe.JoinUser.Kind.Actionable()  {
    				joinClause(comp, fe.JoinUser.Kind, "example.user AS t1 ON t0.user_id=t1.id")
    				if fe.JoinUser.On != nil {
    					comp.Dirty = true
    					if err := UserCriteriaWhereClause(comp, fe.JoinUser.On, 1); err != nil {
    						return "", nil, err
    					}
    				}
    			}
    			if fe.JoinWpis != nil && fe.JoinWpis.Kind.Actionable()  {
    				joinClause(comp, fe.JoinWpis.Kind, "post AS t2 ON ")
    				if fe.JoinWpis.On != nil {
    					comp.Dirty = true
    					if err := PostCriteriaWhereClause(comp, fe.JoinWpis.On, 2); err != nil {
    						return "", nil, err
    					}
    				}
    			}
    			if comp.Dirty {
    				buf.ReadFrom(comp)
    				comp.Dirty = false
    			}
    			if fe.Where != nil {
    				if err := CommentCriteriaWhereClause(comp, fe.Where, 0); err != nil {
    					return "", nil, err
    				}
    			}
    			if fe.JoinUser != nil && fe.JoinUser.Kind.Actionable() && fe.JoinUser.Where != nil {
    				if err := UserCriteriaWhereClause(comp, fe.JoinUser.Where, 1); err != nil {
    					return "", nil, err
    				}
    			}
    			if fe.JoinWpis != nil && fe.JoinWpis.Kind.Actionable() && fe.JoinWpis.Where != nil {
    				if err := PostCriteriaWhereClause(comp, fe.JoinWpis.Where, 2); err != nil {
    					return "", nil, err
    				}
    			}
    			if comp.Dirty {
    				if _, err := buf.WriteString(" WHERE "); err != nil {
    					return "", nil, err
    				}
    				buf.ReadFrom(comp)
    			}

    			if len(fe.OrderBy) > 0 {
    				i := 0
    				for _, order := range fe.OrderBy {
    					for _, columnName := range TableCommentColumns {
    						if order.Name == columnName {
    							if i == 0 {
    								comp.WriteString(" ORDER BY ")
    							}
    							if i > 0 {
    								if _, err := comp.WriteString(", "); err != nil {
    									return "", nil, err
    								}
    							}
    							if _, err := comp.WriteString(order.Name); err != nil {
    								return "", nil, err
    							}
    							if order.Descending {
    								if _, err := comp.WriteString(" DESC"); err != nil {
    									return "", nil, err
    								}
    							}
    							i++
    							break
    						}
    					}
    				}
    			}
    			if fe.Offset > 0 {
    				if _, err := comp.WriteString(" OFFSET "); err != nil {
    					return "", nil, err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				if _, err := comp.WriteString(" "); err != nil {
    					return "", nil, err
    				}
    				comp.Add(fe.Offset)
    			}
    			if fe.Limit > 0 {
    				if _, err := comp.WriteString(" LIMIT "); err != nil {
    					return "", nil, err
    				}
    				if err := comp.WritePlaceholder(); err != nil {
    					return "", nil, err
    				}
    				if _, err := comp.WriteString(" "); err != nil {
    					return "", nil, err
    				}
    				comp.Add(fe.Limit)
    			}

    			buf.ReadFrom(comp)

    			return buf.String(), comp.Args(), nil
    		}

    		func (r *CommentRepositoryBase) Find(ctx context.Context, fe *CommentFindExpr) ([]*CommentEntity, error) {
    			query, args, err := r.FindQuery(fe)
    			if err != nil {
    				return nil, err
    			}
    			rows, err := r.DB.QueryContext(ctx, query, args...)
    			if r.Log != nil {
    				r.Log(err, TableComment, "find", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			defer rows.Close()
    			var (
					entities []*CommentEntity
    				props []interface{}
				)
    			for rows.Next() {
    				var ent CommentEntity
    				if props, err = ent.Props(); err != nil {
    					return nil, err
    				}
    				var prop []interface{}
    				if fe.JoinUser != nil && fe.JoinUser.Kind.Actionable() && fe.JoinUser.Fetch {
    					ent.User = &UserEntity{}
    					if prop, err = ent.User.Props(); err != nil {
    						return nil, err
    					}
    					props = append(props, prop...)
    				}
    				if fe.JoinWpis != nil && fe.JoinWpis.Kind.Actionable() && fe.JoinWpis.Fetch {
    					ent.Wpis = &PostEntity{}
    					if prop, err = ent.Wpis.Props(); err != nil {
    						return nil, err
    					}
    					props = append(props, prop...)
    				}
    				err = rows.Scan(props...)
    				if err != nil {
    					return nil, err
    				}

    				entities = append(entities, &ent)
    			}
    			err = rows.Err()
    			if r.Log != nil {
    				r.Log(err, TableComment, "find", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return entities, nil
    		}

    		func (r *CommentRepositoryBase) FindIter(ctx context.Context, fe *CommentFindExpr) (*CommentIterator, error) {
    			query, args, err := r.FindQuery(fe)
    			if err != nil {
    				return nil, err
    			}
    			rows, err := r.DB.QueryContext(ctx, query, args...)
    			if r.Log != nil {
    				r.Log(err, TableComment, "find iter", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return &CommentIterator{
    				rows: rows,
    				expr: fe,
    				cols: fe.Columns,
    			}, nil
    		}

    		func (r *CommentRepositoryBase) UpsertQuery(e *CommentEntity, p *CommentPatch, inf ...string) (string, []interface{}, error) {
    			upsert := NewComposer(2)
    			columns := bytes.NewBuffer(nil)
    			buf := bytes.NewBufferString("INSERT INTO ")
    			buf.WriteString(r.Table)

    			if e.UserID.Valid {
    				if columns.Len() > 0 {
    					if _, err := columns.WriteString(", "); err != nil {
    						return "", nil, err
    					}
    				}
    				if _, err := columns.WriteString(TableCommentColumnUserID); err != nil {
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
    				upsert.Add(e.UserID)
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
    				if p.UserID.Valid {
    					if upsert.Dirty {
    						if _, err := upsert.WriteString(", "); err != nil {
    							return "", nil, err
    						}
    					}
    					if _, err := upsert.WriteString(TableCommentColumnUserID); err != nil {
    						return "", nil, err
    					}
    					if _, err := upsert.WriteString("="); err != nil {
    						return "", nil, err
    					}
    					if err := upsert.WritePlaceholder(); err != nil {
    						return "", nil, err
    					}
    					upsert.Add(p.UserID)
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
    					buf.WriteString("user_id")
    				}
    			}
    			return buf.String(), upsert.Args(), nil
    		}

    		func (r *CommentRepositoryBase) Upsert(ctx context.Context, e *CommentEntity, p *CommentPatch, inf ...string) (*CommentEntity, error) {
    			query, args, err := r.UpsertQuery(e, p, inf...)
    			if err != nil {
    				return nil, err
    			}
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(
					&e.UserID,
				)
    			if r.Log != nil {
    				r.Log(err, TableComment, "upsert", query, args...)
    			}
    			if err != nil {
    				return nil, err
    			}
    			return e, nil
    		}

    		func (r *CommentRepositoryBase) Count(ctx context.Context, c *CommentCountExpr) (int64, error) {
    			query, args, err := r.FindQuery(&CommentFindExpr{
    				Where:   c.Where,
    				Columns: []string{"COUNT(*)"},

    				JoinUser: c.JoinUser,
    				JoinWpis: c.JoinWpis,
    			})
    			if err != nil {
    				return 0, err
    			}
    			var count int64
    			err = r.DB.QueryRowContext(ctx, query, args...).Scan(&count)
    			if r.Log != nil {
    				r.Log(err, TableComment, "count", query, args...)
    			}
    			if err != nil {
    				return 0, err
    			}
    			return count, nil
    		}

    		const (
    			JoinInner = iota
    			JoinLeft
    			JoinRight
    			JoinCross
    			JoinDoNot
    		)

    		type JoinType int

    		func (jt JoinType) String() string {
    			switch jt {

    			case JoinInner:
    				return "INNER JOIN"
    			case JoinLeft:
    				return "LEFT JOIN"
    			case JoinRight:
    				return "RIGHT JOIN"
    			case JoinCross:
    				return "CROSS JOIN"
    			default:
    				return ""
    			}
    		}

			// Actionable returns true if JoinType is one of the known type except JoinDoNot.
			func (jt JoinType) Actionable() bool {
				switch jt {
				case JoinInner, JoinLeft, JoinRight, JoinCross:
					return true
				default:
					return false
				}
			}

    		// ErrorConstraint returns the error constraint of err if it was produced by the pq library.
    		// Otherwise, it returns empty string.
    		func ErrorConstraint(err error) string {
    			if err == nil {
    				return ""
    			}
    			if pqerr, ok := err.(*pq.Error); ok {
    				return pqerr.Constraint
    			}

    			return ""
    		}

    		type RowOrder struct {
    			Name       string
    			Descending bool
    		}

    		type NullInt64Array struct {
    			pq.Int64Array
    			Valid bool
    		}

    		func (n *NullInt64Array) Scan(value interface{}) error {
    			if value == nil {
    				n.Int64Array, n.Valid = nil, false
    				return nil
    			}
    			n.Valid = true
    			return n.Int64Array.Scan(value)
    		}

    		type NullFloat64Array struct {
    			pq.Float64Array
    			Valid bool
    		}

    		func (n *NullFloat64Array) Scan(value interface{}) error {
    			if value == nil {
    				n.Float64Array, n.Valid = nil, false
    				return nil
    			}
    			n.Valid = true
    			return n.Float64Array.Scan(value)
    		}

    		type NullBoolArray struct {
    			pq.BoolArray
    			Valid bool
    		}

    		func (n *NullBoolArray) Scan(value interface{}) error {
    			if value == nil {
    				n.BoolArray, n.Valid = nil, false
    				return nil
    			}
    			n.Valid = true
    			return n.BoolArray.Scan(value)
    		}

    		type NullStringArray struct {
    			pq.StringArray
    			Valid bool
    		}

    		func (n *NullStringArray) Scan(value interface{}) error {
    			if value == nil {
    				n.StringArray, n.Valid = nil, false
    				return nil
    			}
    			n.Valid = true
    			return n.StringArray.Scan(value)
    		}

    		type NullByteaArray struct {
    			pq.ByteaArray
    			Valid bool
    		}

    		func (n *NullByteaArray) Scan(value interface{}) error {
    			if value == nil {
    				n.ByteaArray, n.Valid = nil, false
    				return nil
    			}
    			n.Valid = true
    			return n.ByteaArray.Scan(value)
    		}

    		const (
    			jsonArraySeparator     = ","
    			jsonArrayBeginningChar = "["
    			jsonArrayEndChar       = "]"
    		)

    		// JSONArrayInt64 is a slice of int64s that implements necessary interfaces.
    		type JSONArrayInt64 []int64

    		// Scan satisfy sql.Scanner interface.
    		func (a *JSONArrayInt64) Scan(src interface{}) error {
    			if src == nil {
    				if a == nil {
    					*a = make(JSONArrayInt64, 0)
    				}
    				return nil
    			}

    			var tmp []string
    			var srcs string

    			switch t := src.(type) {
    			case []byte:
    				srcs = string(t)
    			case string:
    				srcs = t
    			default:
    				return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
    			}

    			l := len(srcs)

    			if l < 2 {
    				return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
    			}

    			if l == 2 {
    				*a = make(JSONArrayInt64, 0)
    				return nil
    			}

    			if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
    				return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s", srcs)
    			}

    			tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
    			*a = make(JSONArrayInt64, 0, len(tmp))
    			for i, v := range tmp {
    				j, err := strconv.ParseInt(v, 10, 64)
    				if err != nil {
    					return fmt.Errorf("expected to get source argument in format '[1,2,...,N]', but got %s at index %d", v, i)
    				}

    				*a = append(*a, j)
    			}

    			return nil
    		}

    		// Value satisfy driver.Valuer interface.
    		func (a JSONArrayInt64) Value() (driver.Value, error) {
    			var (
    				buffer bytes.Buffer
    				err    error
    			)

    			if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
    				return nil, err
    			}

    			for i, v := range a {
    				if i > 0 {
    					if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
    						return nil, err
    					}
    				}
    				if _, err := buffer.WriteString(strconv.FormatInt(v, 10)); err != nil {
    					return nil, err
    				}
    			}

    			if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
    				return nil, err
    			}

    			return buffer.Bytes(), nil
    		}

    		// JSONArrayString is a slice of strings that implements necessary interfaces.
    		type JSONArrayString []string

    		// Scan satisfy sql.Scanner interface.
    		func (a *JSONArrayString) Scan(src interface{}) error {
    			if src == nil {
    				if a == nil {
    					*a = make(JSONArrayString, 0)
    				}
    				return nil
    			}

    			switch t := src.(type) {
    			case []byte:
    				return json.Unmarshal(t, a)
    			default:
    				return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
    			}
    		}

    		// Value satisfy driver.Valuer interface.
    		func (a JSONArrayString) Value() (driver.Value, error) {
    			return json.Marshal(a)
    		}

    		// JSONArrayFloat64 is a slice of int64s that implements necessary interfaces.
    		type JSONArrayFloat64 []float64

    		// Scan satisfy sql.Scanner interface.
    		func (a *JSONArrayFloat64) Scan(src interface{}) error {
    			if src == nil {
    				if a == nil {
    					*a = make(JSONArrayFloat64, 0)
    				}
    				return nil
    			}

    			var tmp []string
    			var srcs string

    			switch t := src.(type) {
    			case []byte:
    				srcs = string(t)
    			case string:
    				srcs = t
    			default:
    				return fmt.Errorf("expected slice of bytes or string as a source argument in Scan, not %T", src)
    			}

    			l := len(srcs)

    			if l < 2 {
    				return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
    			}

    			if l == 2 {
    				*a = make(JSONArrayFloat64, 0)
    				return nil
    			}

    			if string(srcs[0]) != jsonArrayBeginningChar || string(srcs[l-1]) != jsonArrayEndChar {
    				return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s", srcs)
    			}

    			tmp = strings.Split(string(srcs[1:l-1]), jsonArraySeparator)
    			*a = make(JSONArrayFloat64, 0, len(tmp))
    			for i, v := range tmp {
    				j, err := strconv.ParseFloat(v, 64)
    				if err != nil {
    					return fmt.Errorf("expected to get source argument in format '[1.3,2.4,...,N.M]', but got %s at index %d", v, i)
    				}

    				*a = append(*a, j)
    			}

    			return nil
    		}

    		// Value satisfy driver.Valuer interface.
    		func (a JSONArrayFloat64) Value() (driver.Value, error) {
    			var (
    				buffer bytes.Buffer
    				err    error
    			)

    			if _, err = buffer.WriteString(jsonArrayBeginningChar); err != nil {
    				return nil, err
    			}

    			for i, v := range a {
    				if i > 0 {
    					if _, err := buffer.WriteString(jsonArraySeparator); err != nil {
    						return nil, err
    					}
    				}
    				if _, err := buffer.WriteString(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
    					return nil, err
    				}
    			}

    			if _, err = buffer.WriteString(jsonArrayEndChar); err != nil {
    				return nil, err
    			}

    			return buffer.Bytes(), nil
    		}

    		var (
    			// Space is a shorthand composition option that holds space.
    			Space = &CompositionOpts{
    				Joint: " ",
    			}
    			// And is a shorthand composition option that holds AND operator.
    			And = &CompositionOpts{
    				Joint: " AND ",
    			}
    			// Or is a shorthand composition option that holds OR operator.
    			Or = &CompositionOpts{
    				Joint: " OR ",
    			}
    			// Comma is a shorthand composition option that holds comma.
    			Comma = &CompositionOpts{
    				Joint: ", ",
    			}
    		)

    		// CompositionOpts is a container for modification that can be applied.
    		type CompositionOpts struct {
    			Joint                           string
    			PlaceholderFuncs, SelectorFuncs []string
    			PlaceholderCast, SelectorCast   string
    			IsJSON                          bool
    			IsDynamic                       bool
    		}

    		// CompositionWriter is a simple wrapper for WriteComposition function.
    		type CompositionWriter interface {
    			// WriteComposition is a function that allow custom struct type to be used as a part of criteria.
    			// It gives possibility to write custom query based on object that implements this interface.
    			WriteComposition(string, *Composer, *CompositionOpts) error
    		}

    		// Composer holds buffer, arguments and placeholders count.
    		// In combination with external buffet can be also used to also generate sub-queries.
    		// To do that simply write buffer to the parent buffer, composer will hold all arguments and remember number of last placeholder.
    		type Composer struct {
    			buf     bytes.Buffer
    			args    []interface{}
    			counter int
    			Dirty   bool
    		}

    		// NewComposer allocates new Composer with inner slice of arguments of given size.
    		func NewComposer(size int64) *Composer {
    			return &Composer{
    				counter: 1,
    				args:    make([]interface{}, 0, size),
    			}
    		}

    		// WriteString appends the contents of s to the query buffer, growing the buffer as
    		// needed. The return value n is the length of s; err is always nil. If the
    		// buffer becomes too large, WriteString will panic with bytes ErrTooLarge.
    		func (c *Composer) WriteString(s string) (int, error) {
    			return c.buf.WriteString(s)
    		}

    		// Write implements io Writer interface.
    		func (c *Composer) Write(b []byte) (int, error) {
    			return c.buf.Write(b)
    		}

    		// Read implements io Reader interface.
    		func (c *Composer) Read(b []byte) (int, error) {
    			return c.buf.Read(b)
    		}

    		// ResetBuf resets internal buffer.
    		func (c *Composer) ResetBuf() {
    			c.buf.Reset()
    		}

    		// String implements fmt Stringer interface.
    		func (c *Composer) String() string {
    			return c.buf.String()
    		}

    		// WritePlaceholder writes appropriate placeholder to the query buffer based on current state of the composer.
    		func (c *Composer) WritePlaceholder() error {
    			if _, err := c.buf.WriteString("$"); err != nil {
    				return err
    			}
    			if _, err := c.buf.WriteString(strconv.Itoa(c.counter)); err != nil {
    				return err
    			}

    			c.counter++
    			return nil
    		}

    		func (c *Composer) WriteAlias(i int) error {
    			if i < 0 {
    				return nil
    			}
    			if _, err := c.buf.WriteString("t"); err != nil {
    				return err
    			}
    			if _, err := c.buf.WriteString(strconv.Itoa(i)); err != nil {
    				return err
    			}
    			if _, err := c.buf.WriteString("."); err != nil {
    				return err
    			}
    			return nil
    		}

    		// Len returns number of arguments.
    		func (c *Composer) Len() int {
    			return c.counter
    		}

    		// Add appends list with new element.
    		func (c *Composer) Add(arg interface{}) {
    			c.args = append(c.args, arg)
    		}

    		// Args returns all arguments stored as a slice.
    		func (c *Composer) Args() []interface{} {
    			return c.args
    		}`
