package pqt

import "fmt"

// Type is a common interface that needs to be implemented so a type can be considered the Type in PQT sense.
type Type interface {
	fmt.Stringer
	// Fingerprint returns unique identifier of the type. Two different types can have same SQL representation.
	Fingerprint() string
}

// BaseType ...
type BaseType struct {
	name string
}

// String implements Stringer interface.
func (bt BaseType) String() string {
	return bt.name
}

// Fingerprint implements Type interface.
func (bt BaseType) Fingerprint() string {
	return fmt.Sprintf("base: %s", bt.name)
}

// TypeDecimal ...
func TypeDecimal(precision, scale int) BaseType {
	switch {
	case precision == 0:
		return BaseType{name: "DECIMAL"}
	case precision != 0 && scale == 0:
		return BaseType{name: fmt.Sprintf("DECIMAL(%d)", precision)}
	default:
		return BaseType{name: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)}
	}
}

// TypeReal represents single precision floating-point numbers.
// In postgres it is stored as 4-byte single-precision floating point numbers.
func TypeReal() BaseType {
	return BaseType{name: "REAL"}
}

// TypeSerial is an auto-incrementing integer.
// It is generally used to store the primary key of a table.
// To specify that a column is to be used as a serial column, declare it as type SERIAL.
// Note that, even though SERIAL appears to be a column type,
// it is actually shorthand notation that tells PostgreSQL to create a auto-incrementing column behind the scenes.
func TypeSerial() BaseType {
	return BaseType{name: "SERIAL"}
}

// TypeSerialSmall is an auto-incrementing small integer.
// It is generally used to store the primary key of a table.
// To specify that a column is to be used as a serial column, declare it as type SMALLSERIAL.
// Note that, even though SMALLSERIAL appears to be a column type,
// it is actually shorthand notation that tells PostgreSQL to create a auto-incrementing column behind the scenes.
func TypeSerialSmall() BaseType {
	return BaseType{name: "SMALLSERIAL"}
}

// TypeSerialBig is an auto-incrementing big integer.
// It is generally used to store the primary key of a table.
// To specify that a column is to be used as a serial column, declare it as type BIGSERIAL.
// Note that, even though BIGSERIAL appears to be a column type,
// it is actually shorthand notation that tells PostgreSQL to create a auto-incrementing column behind the scenes.
func TypeSerialBig() BaseType {
	return BaseType{name: "BIGSERIAL"}
}

// TypeInteger is the common choice, as it offers the best balance between range, storage size, and performance.
func TypeInteger() BaseType {
	return BaseType{name: "INTEGER"}
}

// TypeIntegerSmall is generally only used if disk space is at a premium.
func TypeIntegerSmall() BaseType {
	return BaseType{name: "SMALLINT"}
}

// TypeIntegerBig is designed to be used when the range of the TypeInteger is insufficient.
func TypeIntegerBig() BaseType {
	return BaseType{name: "BIGINT"}
}

// TypeIntegerArray is an array of integers.
func TypeIntegerArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "INTEGER[]"}
	}
	return BaseType{name: fmt.Sprintf("INTEGER[%d]", l)}
}

// TypeIntegerBigArray is an array of big integers.
func TypeIntegerBigArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "BIGINT[]"}
	}
	return BaseType{name: fmt.Sprintf("BIGINT[%d]", l)}
}

// TypeIntegerSmallArray is an array of small integers.
func TypeIntegerSmallArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "SMALLINT[]"}
	}
	return BaseType{name: fmt.Sprintf("SMALLINT[%d]", l)}
}

// TypeDoubleArray is an array of double precision floating-point numbers.
func TypeDoubleArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "DOUBLE PRECISION[]"}
	}
	return BaseType{name: fmt.Sprintf("DOUBLE PRECISION[%d]", l)}
}

// TypeNumeric can store numbers with a very large number of digits.
// It is especially recommended for storing monetary amounts and other quantities where exactness is required.
// Calculations with numeric values yield exact results where possible, e.g. addition, subtraction, multiplication.
// However, calculations on numeric values are very slow compared to the integer types, or to the floating-point types described in the next section.
func TypeNumeric(precision, scale int) BaseType {
	switch {
	case precision == 0:
		return BaseType{name: "NUMERIC"}
	case precision != 0 && scale == 0:
		return BaseType{name: fmt.Sprintf("NUMERIC(%d)", precision)}
	default:
		return BaseType{name: fmt.Sprintf("NUMERIC(%d,%d)", precision, scale)}
	}
}

// TypeDoublePrecision is a numeric type with 15 decimal digits precision.
func TypeDoublePrecision() BaseType {
	return BaseType{name: "DOUBLE PRECISION"}
}

// TypeBool is a state of true or false.
func TypeBool() BaseType {
	return BaseType{name: "BOOL"}
}

// TypeUUID stores Universally Unique Identifiers (UUID) as defined by RFC 4122, ISO/IEC 9834-8:2005, and related standards.
// (Some systems refer to this data type as a globally unique identifier, or GUID, instead.)
// This identifier is a 128-bit quantity that is generated by an algorithm chosen to make it very unlikely that the same identifier will be generated by anyone else in the known universe using the same algorithm.
// Therefore, for distributed systems, these identifiers provide a better uniqueness guarantee than sequence generators, which are only unique within a single database.
func TypeUUID() BaseType {
	return BaseType{name: "UUID"}
}

// TypeCharacter is physically padded with spaces to the specified width n, and are stored and displayed that way.
func TypeCharacter(l int) BaseType {
	return BaseType{name: fmt.Sprintf("CHARACTER[%d]", l)}
}

// TypeText is variable-length character string.
func TypeText() BaseType {
	return BaseType{name: "TEXT"}
}

// TypeTextArray is an array of text.
func TypeTextArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "TEXT[]"}
	}
	return BaseType{name: fmt.Sprintf("TEXT[%d]", l)}
}

// TypeVarchar is a character varying(n), where n is a positive integer.
func TypeVarchar(l int) BaseType {
	if l == 0 {
		return BaseType{name: "VARCHAR"}

	}
	return BaseType{name: fmt.Sprintf("VARCHAR(%d)", l)}
}

// TypeBytea is a binary string.
func TypeBytea() BaseType {
	return BaseType{name: "BYTEA"}
}

// TypeTimestamp is a date and time (no time zone).
func TypeTimestamp() BaseType {
	return BaseType{name: "TIMESTAMP"}
}

// TypeTimestampTZ is a date and time, including time zone
func TypeTimestampTZ() BaseType {
	return BaseType{name: "TIMESTAMPTZ"}
}

// TypeDate is a date only (no time, no time zone).
func TypeDate() BaseType {
	return BaseType{name: "DATE"}
}

// TypeJSON is for storing JSON (JavaScript Object Notation) data, as specified in RFC 7159.
// Such data can also be stored as text, but the JSON data types have the advantage of enforcing that each stored value is valid according to the JSON rules.
func TypeJSON() BaseType {
	return BaseType{name: "JSON"}
}

// TypeJSONB in compare to TypeJSON is stored in a decomposed binary format that makes it slightly slower to input due to added conversion overhead, but significantly faster to process, since no reparsing is needed.
// JSONB also supports indexing, which can be a significant advantage.
func TypeJSONB() BaseType {
	return BaseType{name: "JSONB"}
}

// CompositeType represents the structure of a row or record.
// It is essentially just a list of field names and their data types.
// PostgreSQL allows composite types to be used in many of the same ways that simple types can be used.
// For example, a column of a table can be declared to be of a composite type.
// EXPERIMENTAL
type CompositeType struct {
	name       string
	Attributes []*Attribute
}

// String implements Stringer interface.
func (ct CompositeType) String() string {
	return "" // TODO: ?
}

// Fingerprint implements Type interface.
func (ct CompositeType) Fingerprint() string {
	return fmt.Sprintf("composite: %v", ct)
}

// TypeComposite allocates CompositeType with given name and attributes.
func TypeComposite(name string, attributes ...*Attribute) CompositeType {
	return CompositeType{
		name:       name,
		Attributes: attributes,
	}
}

// EnumeratedType is a data type consisting of a set of named values called elements, members, enumeral, or enumerators of the type.
// EXPERIMENTAL
type EnumeratedType struct {
	name  string
	Enums []string
}

// String implements Stringer interface.
func (et EnumeratedType) String() string {
	return et.name
}

// Fingerprint implements Type interface.
func (et EnumeratedType) Fingerprint() string {
	return fmt.Sprintf("enumarated: %v", et)
}

// TypeEnumerated initializes EnumeratedType with given name and enums.
func TypeEnumerated(name string, enums ...string) EnumeratedType {
	return EnumeratedType{
		name:  name,
		Enums: enums,
	}
}

// PseudoType ...
// EXPERIMENTAL
type PseudoType struct {
	name string
	//	input, output Function
}

// String implements Stringer interface.
func (pt PseudoType) String() string {
	return pt.name
}

// Fingerprint implements Type interface.
func (pt PseudoType) Fingerprint() string {
	return fmt.Sprintf("pseudo: %v", pt)
}

// TypePseudo initializes PseudoType with given name.
func TypePseudo(name string) PseudoType {
	return PseudoType{
		name: name,
	}
}

// MappableType is a type that can be mapped to other types.
// It allows
type MappableType struct {
	From    Type
	Mapping []Type
}

// String implements Stringer interface.
func (mt MappableType) String() string {
	return mt.From.String()
}

// Fingerprint implements Type interface.
func (mt MappableType) Fingerprint() string {
	return fmt.Sprintf("mappable: %v", mt)
}

// TypeMappable ...
func TypeMappable(from Type, mapping ...Type) MappableType {
	return MappableType{
		From:    from,
		Mapping: mapping,
	}
}
