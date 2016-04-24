package pqtgo

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"sort"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/piotrkowalczuk/pqt"
)

const (
	modeDefault = iota
	modeMandatory
	modeOptional
	modeCriteria
)

// Generator ...
type Generator struct {
	acronyms map[string]string
	imports  []string
	pkg      string
}

// NewGenerator ...
func NewGenerator() *Generator {
	return &Generator{
		pkg: "main",
	}
}

// SetAcronyms ...
func (g *Generator) SetAcronyms(acronyms map[string]string) *Generator {
	g.acronyms = acronyms

	return g
}

// SetImports ...
func (g *Generator) SetImports(imports ...string) *Generator {
	g.imports = imports

	return g
}

// AddImport ...
func (g *Generator) AddImport(i string) *Generator {
	if g.imports == nil {
		g.imports = make([]string, 0, 1)
	}

	g.imports = append(g.imports, i)
	return g
}

// SetPackage ...
func (g *Generator) SetPackage(pkg string) *Generator {
	g.pkg = pkg

	return g
}

// Generate ...
func (g *Generator) Generate(s *pqt.Schema) ([]byte, error) {
	code, err := g.generate(s)
	if err != nil {
		return nil, err
	}

	return code.Bytes(), nil
}

// GenerateTo ...
func (g *Generator) GenerateTo(s *pqt.Schema, w io.Writer) error {
	code, err := g.generate(s)
	if err != nil {
		return err
	}

	_, err = code.WriteTo(w)
	return err
}

func (g *Generator) generate(s *pqt.Schema) (*bytes.Buffer, error) {
	code := bytes.NewBuffer(nil)

	g.generatePackage(code)
	g.generateImports(code, s)
	for _, table := range s.Tables {
		g.generateConstants(code, table)
		g.generateColumns(code, table)
		g.generateEntity(code, table)
		g.generateCriteria(code, table)
		g.generateRepository(code, table)
	}

	return code, nil
}

func (g *Generator) generatePackage(code *bytes.Buffer) {
	fmt.Fprintf(code, "package %s \n", g.pkg)
}

func (g *Generator) generateImports(code *bytes.Buffer, schema *pqt.Schema) {
	imports := []string{}

	for _, t := range schema.Tables {
		for _, c := range t.Columns {
			if ct, ok := c.Type.(CustomType); ok {
				imports = append(imports, ct.pkg)
			}
		}
	}

	code.WriteString("import (\n")
	for _, imp := range imports {
		fmt.Fprintf(code, `"%s"`, imp)
	}
	code.WriteString(")\n")
}

func (g *Generator) generateEntity(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("type " + g.private(table.Name) + "Entity struct {")
	for _, c := range table.Columns {
		code.WriteString(g.public(c.Name))
		code.WriteRune(' ')
		g.generateType(code, c, 0)
		code.WriteRune('\n')
	}
	for _, r := range table.OwnedRelationships {
		switch r.Type {
		case pqt.RelationshipTypeOneToMany:
			if r.InversedName != "" {
				code.WriteString(g.public(r.InversedName))
			} else {
				code.WriteString(g.public(r.InversedTable.Name) + "s")
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "[]*%sEntity", g.private(r.InversedTable.Name))
			code.WriteRune('\n')
		case pqt.RelationshipTypeOneToOne, pqt.RelationshipTypeManyToOne:
			if r.InversedName != "" {
				code.WriteString(g.public(r.InversedName))
			} else {
				code.WriteString(g.public(r.InversedTable.Name))
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "*%sEntity", g.private(r.InversedTable.Name))
			code.WriteRune('\n')
		case pqt.RelationshipTypeManyToMany:
			if r.OwnerName != "" {
				code.WriteString(g.public(r.OwnerName))
			} else {
				code.WriteString(g.public(r.OwnerTable.Name))
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "*%sEntity", g.private(r.OwnerTable.Name))
			code.WriteRune('\n')

			if r.InversedName != "" {
				code.WriteString(g.public(r.InversedName))
			} else {
				code.WriteString(g.public(r.InversedTable.Name))
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "*%sEntity", g.private(r.InversedTable.Name))
			code.WriteRune('\n')
		}
	}
	for _, r := range table.InversedRelationships {
		switch r.Type {
		case pqt.RelationshipTypeOneToMany:
			if r.OwnerName != "" {
				code.WriteString(g.public(r.OwnerName))
			} else {
				code.WriteString(g.public(r.OwnerTable.Name))
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "*%sEntity", g.private(r.OwnerTable.Name))
			code.WriteRune('\n')
		case pqt.RelationshipTypeOneToOne, pqt.RelationshipTypeManyToOne:
			if r.OwnerName != "" {
				code.WriteString(g.public(r.OwnerName))
			} else {
				code.WriteString(g.public(r.OwnerTable.Name) + "s")
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "[]*%sEntity", g.private(r.OwnerTable.Name))
			code.WriteRune('\n')
		}
	}
	for _, r := range table.ManyToManyRelationships {
		if r.Type != pqt.RelationshipTypeManyToMany {
			continue
		}

		switch {
		case r.OwnerTable == table:
			if r.InversedName != "" {
				code.WriteString(g.public(r.InversedName))
			} else {
				code.WriteString(g.public(r.InversedTable.Name))
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "[]*%sEntity", g.private(r.InversedTable.Name))
			code.WriteRune('\n')
		case r.InversedTable == table:
			if r.OwnerName != "" {
				code.WriteString(g.public(r.OwnerName))
			} else {
				code.WriteString(g.public(r.OwnerTable.Name) + "s")
			}
			code.WriteRune(' ')
			fmt.Fprintf(code, "[]*%sEntity", g.private(r.OwnerTable.Name))
			code.WriteRune('\n')
		}
	}
	code.WriteString("}\n")
}

func (g *Generator) generateCriteria(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("type " + g.private(table.Name) + "Criteria struct {")
	code.WriteString("offset, limit int64\n")
	code.WriteString("sort map[string]bool\n")

ColumnLoop:
	for _, c := range table.Columns {
		if g.shouldBeColumnIgnoredForCriteria(c) {
			continue ColumnLoop
		}

		code.WriteString(g.private(c.Name))
		code.WriteRune(' ')

		switch c.Type {
		case pqt.TypeJSON(), pqt.TypeJSONB():
			code.WriteString("map[string]interface{}")
		default:
			g.generateType(code, c, modeCriteria)
		}
		code.WriteRune('\n')
	}
	code.WriteString("}\n")
}

func (g *Generator) generateType(code *bytes.Buffer, c *pqt.Column, mode int32) {
	code.WriteString(g.generateTypeString(c, mode))
}

func (g *Generator) generateTypeString(col *pqt.Column, mode int32) string {
	var (
		t         string
		mandatory bool
		criteria  bool
	)

	if str, ok := col.Type.(fmt.Stringer); ok {
		t = str.String()
	} else {
		t = "struct{}"
	}

	switch mode {
	case modeCriteria:
		criteria = true
	case modeMandatory:
		mandatory = true
	case modeOptional:
		mandatory = false
	default:
		mandatory = col.NotNull || col.PrimaryKey
	}

	switch tp := col.Type.(type) {
	case pqt.MappableType:
		for _, mapto := range tp.Mapping {
			switch mt := mapto.(type) {
			case BuiltinType:
				t = generateBuiltinType(mt, mandatory, criteria)
			case CustomType:
				t = mt.String()
			}
		}
	case BuiltinType:
		t = generateBuiltinType(tp, mandatory, criteria)
	case pqt.BaseType:
		t = generateBaseType(tp, mandatory, criteria)
	}

	return t
}

func (g *Generator) generateConstants(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("const (\n")
	g.generateConstantsColumns(code, table)
	g.generateConstantsConstraints(code, table)
	code.WriteString(")\n")
}

func (g *Generator) generateConstantsColumns(code *bytes.Buffer, table *pqt.Table) {
	fmt.Fprintf(code, `table%s = "%s"`, g.public(table.Name), table.FullName())
	code.WriteRune('\n')

	for _, name := range sortedColumns(table.Columns) {
		fmt.Fprintf(code, `table%sColumn%s = "%s"`, g.public(table.Name), g.public(name), name)
		code.WriteRune('\n')
	}
}

func (g *Generator) generateConstantsConstraints(code *bytes.Buffer, table *pqt.Table) {
	for _, c := range tableConstraints(table) {
		name := fmt.Sprintf("%s", pqt.JoinColumns(c.Columns, "_"))
		switch c.Type {
		case pqt.ConstraintTypeCheck:
			fmt.Fprintf(code, `table%sConstraint%sCheck = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypePrimaryKey:
			fmt.Fprintf(code, `table%sConstraintPrimaryKey = "%s"`, g.public(table.Name), c.String())
		case pqt.ConstraintTypeForeignKey:
			fmt.Fprintf(code, `table%sConstraint%sForeignKey = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeExclusion:
			fmt.Fprintf(code, `table%sConstraint%sExclusion = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeUnique:
			fmt.Fprintf(code, `table%sConstraint%sUnique = "%s"`, g.public(table.Name), g.public(name), c.String())
		case pqt.ConstraintTypeIndex:
			fmt.Fprintf(code, `table%sConstraint%sIndex = "%s"`, g.public(table.Name), g.public(name), c.String())
		}

		code.WriteRune('\n')
	}
}

func (g *Generator) generateColumns(code *bytes.Buffer, table *pqt.Table) {
	code.WriteString("var (\n")
	code.WriteRune('\n')

	code.WriteString("table")
	code.WriteString(g.public(table.Name))
	code.WriteString("Columns = []string{\n")

	for _, name := range sortedColumns(table.Columns) {
		g.writeTableNameColumnNameTo(code, table.Name, name)
		code.WriteRune(',')
		code.WriteRune('\n')
	}
	code.WriteString("}")
	code.WriteString(")\n")
}

func (g *Generator) generateRepository(code *bytes.Buffer, table *pqt.Table) {
	fmt.Fprintf(code, `
		type %sRepository struct {
			table string
			columns []string
			db *sql.DB
		}
	`, g.private(table.Name))
	g.generateRepositoryFind(code, table)
	g.generateRepositoryFindOneByPrimaryKey(code, table)
	g.generateRepositoryInsert(code, table)
	g.generateRepositoryUpdateByPrimaryKey(code, table)
	g.generateRepositoryDeleteByPrimaryKey(code, table)
}

func (g *Generator) generateRepositoryFind(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.private(table.Name)

	fmt.Fprintf(code, `func (r *%sRepository) Find(c *%sCriteria) ([]*%sEntity, error) {`, entityName, entityName, entityName)
	fmt.Fprintf(code, `
			comp := pqcomp.New(2, 0, 1)
			comp.AddArg(c.offset)
			comp.AddArg(c.limit)

			where := comp.Compose(%d)
	`, len(table.Columns))

ColumnLoop:
	for _, c := range table.Columns {
		if g.shouldBeColumnIgnoredForCriteria(c) {
			continue ColumnLoop
		}

		columnNamePrivate := g.private(c.Name)
		columnNameWithTable := g.columnNameWithTableName(c.Table.Name, c.Name)

		switch g.generateTypeString(c, modeCriteria) {
		case "*protot.QueryTimestamp":
			fmt.Fprintf(code, `
				if c.%s != nil && c.%s.Valid {
					%st1 := c.%s.Value()
					if %st1 != nil {
						%s1, err := ptypes.Timestamp(%st1 )
						if err != nil {
							return nil, err
						}

						switch c.%s.Type {
						case protot.NumericQueryType_NOT_A_NUMBER:
							where.AddExpr(%s, pqcomp.IS, pqcomp.NULL)
						case protot.NumericQueryType_EQUAL:
							where.AddExpr(%s, pqcomp.E, %s1)
						case protot.NumericQueryType_NOT_EQUAL:
							where.AddExpr(%s, pqcomp.NE, %s1)
						case protot.NumericQueryType_GREATER:
							where.AddExpr(%s, pqcomp.GT, %s1)
						case protot.NumericQueryType_GREATER_EQUAL:
							where.AddExpr(%s, pqcomp.GTE, %s1)
						case protot.NumericQueryType_LESS:
							where.AddExpr(%s, pqcomp.LT, %s1)
						case protot.NumericQueryType_LESS_EQUAL:
							where.AddExpr(%s, pqcomp.LTE, %s1)
						case protot.NumericQueryType_IN:
							where.AddExpr(%s, pqcomp.IN, %s1)
						case protot.NumericQueryType_BETWEEN:
							%st2 := c.%s.Values[1]
							if %st2 != nil {
								%s2, err := ptypes.Timestamp(%st2)
								if err != nil {
									return nil, err
								}

								where.AddExpr(%s, pqcomp.GT, %s1)
								where.AddExpr(%s, pqcomp.LT, %s2)
							}
						}
					}
				}
			`,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate,
				columnNameWithTable,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
			)
		case "*protot.QueryInt64":
			fmt.Fprintf(code, `
				if c.%s != nil && c.%s.Valid {
					switch c.%s.Type {
					case protot.NumericQueryType_NOT_A_NUMBER:
						where.AddExpr(%s, pqcomp.IS, pqcomp.NULL)
					case protot.NumericQueryType_EQUAL:
						where.AddExpr(%s, pqcomp.E, c.%s.Value())
					case protot.NumericQueryType_NOT_EQUAL:
						where.AddExpr(%s, pqcomp.NE, c.%s.Value())
					case protot.NumericQueryType_GREATER:
						where.AddExpr(%s, pqcomp.GT, c.%s.Value())
					case protot.NumericQueryType_GREATER_EQUAL:
						where.AddExpr(%s, pqcomp.GTE, c.%s.Value())
					case protot.NumericQueryType_LESS:
						where.AddExpr(%s, pqcomp.LT, c.%s.Value())
					case protot.NumericQueryType_LESS_EQUAL:
						where.AddExpr(%s, pqcomp.LTE, c.%s.Value())
					case protot.NumericQueryType_IN:
						for _, v := range c.%s.Values {
							where.AddExpr(%s, pqcomp.IN, v)
						}
					case protot.NumericQueryType_BETWEEN:
						where.AddExpr(%s, pqcomp.GT, c.%s.Values[0])
						where.AddExpr(%s, pqcomp.LT, c.%s.Values[1])
					}
				}
			`,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate,
				columnNameWithTable,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
				columnNamePrivate,
				columnNameWithTable,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, columnNamePrivate,
			)
		case "*protot.QueryString":
			fmt.Fprintf(code, `
				if c.%s != nil && c.%s.Valid {
					switch c.%s.Type {
					case protot.TextQueryType_NOT_A_TEXT:
						if c.%s.Negation {
							where.AddExpr(%s, pqcomp.IS, pqcomp.NOT_NULL)
						} else {
							where.AddExpr(%s, pqcomp.IS, pqcomp.NULL)
						}
					case protot.TextQueryType_EXACT:
						where.AddExpr(%s, pqcomp.E, c.%s.Value())
					case protot.TextQueryType_SUBSTRING:
						where.AddExpr(%s, pqcomp.LIKE, "%s"+c.%s.Value()+"%s")
					case protot.TextQueryType_HAS_PREFIX:
						where.AddExpr(%s, pqcomp.LIKE, c.%s.Value()+"%s")
					case protot.TextQueryType_HAS_SUFFIX:
						where.AddExpr(%s, pqcomp.LIKE, "%s"+c.%s.Value())
					}
				}
			`,
				columnNamePrivate, columnNamePrivate,
				columnNamePrivate,
				columnNamePrivate,
				columnNameWithTable,
				columnNameWithTable,
				columnNameWithTable, columnNamePrivate,
				columnNameWithTable, "%", columnNamePrivate, "%",
				columnNameWithTable, columnNamePrivate, "%",
				columnNameWithTable, "%", columnNamePrivate,
			)
		default:
			code.WriteString("where.AddExpr(")
			g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
			fmt.Fprintf(code, ", pqcomp.E, c.%s)", g.private(c.Name))
		}
		//fmt.Fprintf(code, "if c.%s.From != nil {\n", g.private(c.Name))
		//code.WriteString("where.AddExpr(")
		//g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
		//fmt.Fprintf(code, ", pqcomp.GT, c.%s.From.Time())", g.private(c.Name))
		//code.WriteString("}\n")
		//
		//fmt.Fprintf(code, "if c.%s.To != nil {\n", g.private(c.Name))
		//code.WriteString("where.AddExpr(")
		//g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
		//fmt.Fprintf(code, ", pqcomp.LT, c.%s.To.Time())", g.private(c.Name))
		//code.WriteString("}\n")
		//default:
		//	code.WriteString("where.AddExpr(")
		//	g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
		//	fmt.Fprintf(code, ", pqcomp.E, c.%s)", g.private(c.Name))
		//}

		code.WriteRune('\n')
	}
	fmt.Fprintf(code, `
	rows, err := findQueryComp(r.db, r.table, comp, where, c.sort, r.columns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*%sEntity
	for rows.Next() {
		var entity %sEntity
		err = rows.Scan(
	`, entityName, entityName)
	for _, c := range table.Columns {
		fmt.Fprintf(code, "&entity.%s,\n", g.public(c.Name))
	}
	fmt.Fprintf(code, `)
			if err != nil {
				return nil, err
			}

			entities = append(entities, &entity)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		return entities, nil
	}
	`)
}

func (g *Generator) generateRepositoryFindOneByPrimaryKey(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.private(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(code, `func (r *%sRepository) FindOneBy%s(%s %s) (*%sEntity, error) {`, entityName, g.public(pk.Name), g.private(pk.Name), g.generateTypeString(pk, modeMandatory), entityName)
	fmt.Fprintf(code, `var (
		query string
		entity %sEntity
	)`, entityName)
	code.WriteRune('\n')
	fmt.Fprintf(code, "query = `SELECT ")
	for i, c := range table.Columns {
		fmt.Fprintf(code, "%s", c.Name)
		if i != len(table.Columns)-1 {
			code.WriteRune(',')
		}
		code.WriteRune('\n')
	}
	fmt.Fprintf(code, " FROM %s WHERE %s = $1`", table.FullName(), pk.Name)

	fmt.Fprintf(code, `
	err := r.db.QueryRow(query, %s).Scan(
	`, g.private(pk.Name))
	for _, c := range table.Columns {
		fmt.Fprintf(code, "&entity.%s,\n", g.public(c.Name))
	}
	fmt.Fprintf(code, `)
		if err != nil {
			return nil, err
		}

		return &entity, nil
	}
	`)
}

func (g *Generator) generateRepositoryInsert(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.private(table.Name)

	fmt.Fprintf(code, `func (r *%sRepository) Insert(e *%sEntity) (*%sEntity, error) {`, entityName, entityName, entityName)
	fmt.Fprintf(code, `
			insert := pqcomp.New(0, %d)
	`, len(table.Columns))

ColumnsLoop:
	for _, c := range table.Columns {
		// If column has default value for insert. Value will be provided by postgres.
		//if _, ok := c.DefaultOn(pqt.EventInsert); !ok {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue ColumnsLoop
		default:
			code.WriteString("insert.AddExpr(")
			g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
			fmt.Fprintf(code, `, "", e.%s)`, g.public(c.Name))
			code.WriteRune('\n')
		}
		//}
	}
	fmt.Fprintf(code, `err := insertQueryComp(r.db, r.table, insert, r.columns).Scan(`)

	for _, c := range table.Columns {
		fmt.Fprintf(code, "&e.%s,\n", g.public(c.Name))
	}
	fmt.Fprintf(code, `)
		if err != nil {
			return nil, err
		}

		return e, nil
	}
	`)
}

func (g *Generator) generateRepositoryUpdateByPrimaryKey(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.private(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(code, `func (r *%sRepository) UpdateBy%s(`, entityName, g.public(pk.Name))
	code.WriteRune('\n')
	fmt.Fprintf(code, `%s %s,`, g.private(pk.Name), g.generateTypeString(pk, modeMandatory))
	code.WriteRune('\n')

ArgumentsLoop:
	for _, c := range table.Columns {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue ArgumentsLoop
		default:
			fmt.Fprintf(code, `%s %s,`, g.private(c.Name), g.generateTypeString(c, modeOptional))
			code.WriteRune('\n')
		}
	}
	fmt.Fprintf(code, `) (*%sEntity, error) {`, entityName)
	fmt.Fprintf(code, `
			update := pqcomp.New(0, %d)
	`, len(table.Columns))

	code.WriteString("update.AddExpr(")
	g.writeTableNameColumnNameTo(code, table.Name, pk.Name)
	fmt.Fprintf(code, `, pqcomp.E, %s)`, g.private(pk.Name))
	code.WriteRune('\n')

ColumnsLoop:
	for _, c := range table.Columns {
		if c == pk {
			continue ColumnsLoop
		}
		if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprintf(code, `if %s != nil {`, g.private(c.Name))

			}
			code.WriteRune('\n')
		}

		code.WriteString("update.AddExpr(")
		g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
		fmt.Fprintf(code, `, pqcomp.E, %s)`, g.private(c.Name))
		code.WriteRune('\n')

		if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprintf(code, `} else {`)
				code.WriteString("update.AddExpr(")
				g.writeTableNameColumnNameTo(code, c.Table.Name, c.Name)
				fmt.Fprintf(code, `, pqcomp.E, "%s")`, d)
			}
		}
		if _, ok := c.DefaultOn(pqt.EventInsert, pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				code.WriteRune('\n')
				code.WriteString(`}`)
				code.WriteRune('\n')
			}
		}
	}
	fmt.Fprintf(code, `
	if update.Len() == 0 {
		return nil, errors.New("%s: %s update failure, nothing to update")
	}`, g.pkg, entityName)

	fmt.Fprintf(code, `
	query := "UPDATE %s SET "
	for update.Next() {
		if !update.First() {
			query += ", "
		}

		query += update.Key() + " " + update.Oper() + " " + update.PlaceHolder()
	}
	query += " WHERE %s = $1 RETURNING " + strings.Join(r.columns, ", ")
	var e %sEntity
	err := r.db.QueryRow(query, update.Args()...).Scan(
	`, table.FullName(), pk.Name, entityName)
	for _, c := range table.Columns {
		fmt.Fprintf(code, "&e.%s,\n", g.public(c.Name))
	}
	fmt.Fprintf(code, `)
			if err != nil {
				return nil, err
			}


		return &e, nil
	}
	`)
}

func (g *Generator) generateRepositoryDeleteByPrimaryKey(code *bytes.Buffer, table *pqt.Table) {
	entityName := g.private(table.Name)
	pk, ok := table.PrimaryKey()
	if !ok {
		return
	}

	fmt.Fprintf(code, `
		func (r *%sRepository) DeleteBy%s(%s %s) (int64, error) {
			query := "DELETE FROM %s WHERE %s = $1"

			res, err := r.db.Exec(query, %s)
			if err != nil {
				return 0, err
			}

			return res.RowsAffected()
		}
	`, entityName, g.public(pk.Name), g.private(pk.Name), g.generateTypeString(pk, 1), table.FullName(), pk.Name, g.private(pk.Name))
}

func sortedColumns(columns []*pqt.Column) []string {
	tmp := make([]string, 0, len(columns))
	for _, c := range columns {
		tmp = append(tmp, c.Name)
	}
	sort.Strings(tmp)

	return tmp
}

func snake(s string, private bool, acronyms map[string]string) string {
	var parts []string
	parts1 := strings.Split(s, "_")
	for _, p1 := range parts1 {
		parts2 := strings.Split(p1, "/")
		for _, p2 := range parts2 {
			parts3 := strings.Split(p2, "-")
			parts = append(parts, parts3...)
		}
	}

	for i, part := range parts {
		if !private || i > 0 {
			if formatted, ok := acronyms[part]; ok {
				parts[i] = formatted

				continue
			}
		}

		parts[i] = xstrings.FirstRuneToUpper(part)
	}

	if private {
		parts[0] = xstrings.FirstRuneToLower(parts[0])
	}

	return strings.Join(parts, "")
}

func (g *Generator) private(s string) string {
	return snake(s, true, g.acronyms)
}

func (g *Generator) public(s string) string {
	return snake(s, false, g.acronyms)
}

func chooseType(typeMandatory, typeOptional, typeCriteria string, mandatory, criteria bool) string {
	switch {
	case criteria:
		return typeCriteria
	case mandatory:
		return typeMandatory
	default:
		return typeOptional
	}
}

func tableConstraints(t *pqt.Table) []*pqt.Constraint {
	var constraints []*pqt.Constraint
	for _, c := range t.Columns {
		constraints = append(constraints, c.Constraints()...)
	}

	return append(constraints, t.Constraints...)
}

func generateBaseType(t pqt.Type, mandatory, criteria bool) string {
	switch t {
	case pqt.TypeText():
		return chooseType("string", "*nilt.String", "*protot.QueryString", mandatory, criteria)
	case pqt.TypeBool():
		return chooseType("bool", "*nilt.Bool", "*nilt.Bool", mandatory, criteria)
	case pqt.TypeIntegerSmall():
		return "int16"
	case pqt.TypeInteger():
		return chooseType("int32", "*nilt.Int32", "*nilt.Int32", mandatory, criteria)
	case pqt.TypeIntegerBig():
		return chooseType("int64", "*nilt.Int64", "*protot.QueryInt64", mandatory, criteria)
	case pqt.TypeSerial():
		return chooseType("int32", "*nilt.Int32", "*nilt.Int32", mandatory, criteria)
	case pqt.TypeSerialSmall():
		return "int16" // TODO: missing nilt.Int16 type
	case pqt.TypeSerialBig():
		return chooseType("int64", "*nilt.Int64", "*protot.QueryInt64", mandatory, criteria)
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		return chooseType("time.Time", "*time.Time", "*protot.QueryTimestamp", mandatory, criteria)
	case pqt.TypeReal():
		return chooseType("float32", "*nilt.Float32", "*nilt.Float32", mandatory, criteria)
	case pqt.TypeDoublePrecision():
		return chooseType("float64", "*nilt.Float64", "*protot.QueryFloat64", mandatory, criteria)
	case pqt.TypeBytea():
		return "[]byte"
	default:
		gt := t.String()
		switch {
		case strings.HasPrefix(gt, "SMALLINT["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "INTEGER["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "BIGINT["):
			return "pqt.ArrayInt64"
		case strings.HasPrefix(gt, "TEXT["):
			return "pqt.ArrayString"
		case strings.HasPrefix(gt, "DECIMAL"), strings.HasPrefix(gt, "NUMERIC"):
			return chooseType("float32", "*nilt.Float32", "*nilt.Float32", mandatory, criteria)
		case strings.HasPrefix(gt, "VARCHAR"):
			return chooseType("string", "*nilt.String", "*protot.QueryString", mandatory, criteria)
		default:
			return "struct{}"
		}
	}
}

func generateBuiltinType(t BuiltinType, mandatory, criteria bool) (r string) {
	switch types.BasicKind(t) {
	case types.Bool:
		r = chooseType("bool", "*nilt.Bool", "*nilt.Bool", mandatory, criteria)
	case types.Int:
		r = chooseType("int", "*nilt.Int", "*nilt.Int", mandatory, criteria)
	case types.Int8:
		r = chooseType("int8", "*int8", "*int8", mandatory, criteria)
	case types.Int16:
		r = chooseType("int16", "*int16", "*int16", mandatory, criteria)
	case types.Int32:
		r = chooseType("int32", "*nilt.Int32", "*nilt.Int32", mandatory, criteria)
	case types.Int64:
		r = chooseType("int64", "*nilt.Int64", "*protot.QueryInt64", mandatory, criteria)
	case types.Uint:
		r = chooseType("uint", "*uint", "*uint", mandatory, criteria)
	case types.Uint8:
		r = chooseType("uint8", "*uint8", "*uint8", mandatory, criteria)
	case types.Uint16:
		r = chooseType("uint16", "*uint16", "*uint16", mandatory, criteria)
	case types.Uint32:
		r = chooseType("uint32", "*nilt.Uint32", "*nilt.Uint32", mandatory, criteria)
	case types.Uint64:
		r = chooseType("uint64", "*uint64", "*uint64", mandatory, criteria)
	case types.Float32:
		r = chooseType("float32", "*nilt.Float32", "*nilt.Float32", mandatory, criteria)
	case types.Float64:
		r = chooseType("float64", "*nilt.Float64", "*protot.QueryFloat64", mandatory, criteria)
	case types.Complex64:
		r = chooseType("complex64", "*complex64", "*complex64", mandatory, criteria)
	case types.Complex128:
		r = chooseType("complex128", "*complex128", "*complex128", mandatory, criteria)
	case types.String:
		r = chooseType("string", "*nilt.String", "*protot.QueryString", mandatory, criteria)
	default:
		r = "invalid"
	}

	return
}

func (g *Generator) writeTableNameColumnNameTo(w io.Writer, tableName, columnName string) {
	fmt.Fprintf(w, "table%sColumn%s", g.public(tableName), g.public(columnName))
}

func (g *Generator) columnNameWithTableName(tableName, columnName string) string {
	return fmt.Sprintf("table%sColumn%s", g.public(tableName), g.public(columnName))
}

func (g *Generator) shouldBeColumnIgnoredForCriteria(c *pqt.Column) bool {
	if mt, ok := c.Type.(pqt.MappableType); ok {
		switch mt.From {
		case pqt.TypeJSON(), pqt.TypeJSONB():
			return true
		}
	}

	return false
}
