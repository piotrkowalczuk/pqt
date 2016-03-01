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
		case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
			code.WriteString("protot.TimestampRange")
		case pqt.TypeJSON(), pqt.TypeJSONB():
			code.WriteString("map[string]interface{}")
		default:
			g.generateType(code, c, -1)
		}
		code.WriteRune('\n')
	}
	code.WriteString("}\n")
}

func (g *Generator) generateType(code *bytes.Buffer, c *pqt.Column, m int32) {
	code.WriteString(g.generateTypeString(c, m))
}

func (g *Generator) generateTypeString(c *pqt.Column, m int32) string {
	var t string

	if str, ok := c.Type.(fmt.Stringer); ok {
		t = str.String()
	} else {
		t = "struct{}"
	}

	var mandatory bool
	switch m {
	case 1:
		mandatory = true
	case -1:
		mandatory = false
	default:
		mandatory = c.NotNull || c.PrimaryKey
	}

	switch tp := c.Type.(type) {
	case pqt.MappableType:
		for _, mapto := range tp.Mapping {
			switch mt := mapto.(type) {
			case BuiltinType:
				t = generateBuiltinType(mt, mandatory)
			case CustomType:
				t = mt.String()
			}
		}
	case BuiltinType:
		t = generateBuiltinType(tp, mandatory)
	case pqt.BaseType:
		t = generateBaseType(tp, mandatory)
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
		g.writeColumnNameConstraintTo(code, table.Name, name)
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

		switch c.Type {
		case pqt.TypeTimestampTZ(), pqt.TypeTimestamp():
			fmt.Fprintf(code, "if c.%s.From != nil {\n", g.private(c.Name))
			code.WriteString("where.AddExpr(")
			g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
			fmt.Fprintf(code, ", pqcomp.GT, c.%s.From.Time())", g.private(c.Name))
			code.WriteString("}\n")

			fmt.Fprintf(code, "if c.%s.To != nil {\n", g.private(c.Name))
			code.WriteString("where.AddExpr(")
			g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
			fmt.Fprintf(code, ", pqcomp.LT, c.%s.To.Time())", g.private(c.Name))
			code.WriteString("}\n")
		default:
			code.WriteString("where.AddExpr(")
			g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
			fmt.Fprintf(code, ", pqcomp.E, c.%s)", g.private(c.Name))
		}

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

	fmt.Fprintf(code, `func (r *%sRepository) FindOneBy%s(%s %s) (*%sEntity, error) {`, entityName, g.public(pk.Name), g.private(pk.Name), g.generateTypeString(pk, 1), entityName)
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
		if _, ok := c.DefaultOn(pqt.EventInsert); !ok {
			switch c.Type {
			case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
				continue ColumnsLoop
			default:
				code.WriteString("insert.AddExpr(")
				g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
				fmt.Fprintf(code, `, "", e.%s)`, g.public(c.Name))
				code.WriteRune('\n')
			}
		}
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
	fmt.Fprintf(code, `%s %s,`, g.private(pk.Name), g.generateTypeString(pk, 1))
	code.WriteRune('\n')

ArgumentsLoop:
	for _, c := range table.Columns {
		switch c.Type {
		case pqt.TypeSerial(), pqt.TypeSerialBig(), pqt.TypeSerialSmall():
			continue ArgumentsLoop
		default:
			fmt.Fprintf(code, `%s %s,`, g.private(c.Name), g.generateTypeString(c, -1))
			code.WriteRune('\n')
		}
	}
	fmt.Fprintf(code, `) (*%sEntity, error) {`, entityName)
	fmt.Fprintf(code, `
			update := pqcomp.New(0, %d)
	`, len(table.Columns))

	code.WriteString("update.AddExpr(")
	g.writeColumnNameConstraintTo(code, table.Name, pk.Name)
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
		g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
		fmt.Fprintf(code, `, pqcomp.E, %s)`, g.private(c.Name))
		code.WriteRune('\n')

		if d, ok := c.DefaultOn(pqt.EventUpdate); ok {
			switch c.Type {
			case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
				fmt.Fprintf(code, `} else {`)
				code.WriteString("update.AddExpr(")
				g.writeColumnNameConstraintTo(code, c.Table.Name, c.Name)
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

func nullable(typeA, typeB string, mandatory bool) string {
	if mandatory {
		return typeA
	}
	return typeB
}

func tableConstraints(t *pqt.Table) []*pqt.Constraint {
	var constraints []*pqt.Constraint
	for _, c := range t.Columns {
		constraints = append(constraints, c.Constraints()...)
	}

	return append(constraints, t.Constraints...)
}

func generateBaseType(t pqt.Type, mandatory bool) string {
	switch t {
	case pqt.TypeText():
		return nullable("string", "nilt.String", mandatory)
	case pqt.TypeBool():
		return nullable("bool", "nilt.Bool", mandatory)
	case pqt.TypeIntegerSmall():
		return "int16"
	case pqt.TypeInteger():
		return nullable("int32", "nilt.Int32", mandatory)
	case pqt.TypeIntegerBig():
		return nullable("int64", "nilt.Int64", mandatory)
	case pqt.TypeSerial():
		return nullable("int32", "nilt.Int32", mandatory)
	case pqt.TypeSerialSmall():
		return "int16" // TODO: missing nilt.Int16 type
	case pqt.TypeSerialBig():
		return nullable("int64", "nilt.Int64", mandatory)
	case pqt.TypeTimestamp(), pqt.TypeTimestampTZ():
		return nullable("time.Time", "*time.Time", mandatory)
	case pqt.TypeReal():
		return nullable("float32", "nilt.Float32", mandatory)
	case pqt.TypeDoublePrecision():
		return nullable("float64", "nilt.Float64", mandatory)
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
			return nullable("float32", "nilt.Float32", mandatory)
		case strings.HasPrefix(gt, "VARCHAR"):
			return nullable("string", "nilt.String", mandatory)
		default:
			return "struct{}"
		}
	}
}

func generateBuiltinType(t BuiltinType, mandatory bool) (r string) {
	switch types.BasicKind(t) {
	case types.Bool:
		r = nullable("bool", "nilt.Bool", mandatory)
	case types.Int:
		r = nullable("int", "nilt.Int", mandatory)
	case types.Int8:
		r = nullable("int8", "*int8", mandatory)
	case types.Int16:
		r = nullable("int16", "*int16", mandatory)
	case types.Int32:
		r = nullable("int32", "nilt.Int32", mandatory)
	case types.Int64:
		r = nullable("int64", "nilt.Int64", mandatory)
	case types.Uint:
		r = nullable("uint", "*uint", mandatory)
	case types.Uint8:
		r = nullable("uint8", "*uint8", mandatory)
	case types.Uint16:
		r = nullable("uint16", "*uint16", mandatory)
	case types.Uint32:
		r = nullable("uint32", "nilt.Uint32", mandatory)
	case types.Uint64:
		r = nullable("uint64", "*uint64", mandatory)
	case types.Float32:
		r = nullable("float32", "nilt.Float32", mandatory)
	case types.Float64:
		r = nullable("float64", "nilt.Float64", mandatory)
	case types.Complex64:
		r = nullable("complex64", "*complex64", mandatory)
	case types.Complex128:
		r = nullable("complex128", "*complex128", mandatory)
	case types.String:
		r = nullable("string", "nilt.String", mandatory)
	default:
		r = "invalid"
	}

	return
}

func (g *Generator) writeColumnNameConstraintTo(w io.Writer, tableName, columnName string) {
	fmt.Fprintf(w, "table%sColumn%s", g.public(tableName), g.public(columnName))
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
