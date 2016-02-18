package pqt

import "fmt"

// Type ...
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

// TypeReal ...
func TypeReal() BaseType {
	return BaseType{name: "REAL"}
}

// TypeSerial ...
func TypeSerial() BaseType {
	return BaseType{name: "SERIAL"}
}

// TypeSerialSmall ...
func TypeSerialSmall() BaseType {
	return BaseType{name: "SMALLSERIAL"}
}

// TypeSerialBig ...
func TypeSerialBig() BaseType {
	return BaseType{name: "BIGSERIAL"}
}

// TypeInteger ...
func TypeInteger() BaseType {
	return BaseType{name: "INTEGER"}
}

// TypeIntegerSmall ...
func TypeIntegerSmall() BaseType {
	return BaseType{name: "SMALLINT"}
}

// TypeIntegerBig ...
func TypeIntegerBig() BaseType {
	return BaseType{name: "BIGINT"}
}

// TypeIntegerArray ...
func TypeIntegerArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "INTEGER[]"}
	}
	return BaseType{name: fmt.Sprintf("INTEGER[%d]", l)}
}

// TypeIntegerBigArray ...
func TypeIntegerBigArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "BIGINT[]"}
	}
	return BaseType{name: fmt.Sprintf("BIGINT[%d]", l)}
}

// TypeIntegerSmallArray ...
func TypeIntegerSmallArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "SMALLINT[]"}
	}
	return BaseType{name: fmt.Sprintf("SMALLINT[%d]", l)}
}

// TypeNumeric ...
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

// TypeDoublePrecision ...
func TypeDoublePrecision() BaseType {
	return BaseType{name: "DOUBLE PRECISION"}
}

// TypeBool ...
func TypeBool() BaseType {
	return BaseType{name: "BOOL"}
}

// TypeText ...
func TypeText() BaseType {
	return BaseType{name: "TEXT"}
}

// TypeTextArray ...
func TypeTextArray(l int) BaseType {
	if l == 0 {
		return BaseType{name: "TEXT[]"}
	}
	return BaseType{name: fmt.Sprintf("TEXT[%d]", l)}
}

// TypeVarchar ...
func TypeVarchar(l int) BaseType {
	if l == 0 {
		return BaseType{name: "VARCHAR"}

	}
	return BaseType{name: fmt.Sprintf("VARCHAR(%d)", l)}
}

// TypeBytea ...
func TypeBytea() BaseType {
	return BaseType{name: "BYTEA"}
}

// TypeTimestamp ...
func TypeTimestamp() BaseType {
	return BaseType{name: "TIMESTAMP"}
}

// TypeTimestampTZ ...
func TypeTimestampTZ() BaseType {
	return BaseType{name: "TIMESTAMPTZ"}
}

// TypeJSON ...
func TypeJSON() BaseType {
	return BaseType{name: "JSON"}
}

// TypeJSONB ...
func TypeJSONB() BaseType {
	return BaseType{name: "JSONB"}
}

// CompositeType ...
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

// TypeComposite ...
func TypeComposite(name string, attributes ...*Attribute) CompositeType {
	return CompositeType{
		name:       name,
		Attributes: attributes,
	}
}

// EnumeratedType ...
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

// TypeEnumerated ...
func TypeEnumerated(name string, enums ...string) EnumeratedType {
	return EnumeratedType{
		name:  name,
		Enums: enums,
	}
}

// PseudoType ...
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

// TypePseudo ...
func TypePseudo(name string) PseudoType {
	return PseudoType{
		name: name,
	}
}

// MappableType ...
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
