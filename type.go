package pqt

import "fmt"

type Type interface {
	fmt.Stringer
	// Fingerprint returns unique identifier of the type. Two different types can have same SQL representation.
	Fingerprint() string
}

type BaseType struct {
	name string
	//	input, output Function
}

// String implements Stringer interface.
func (bt BaseType) String() string {
	return bt.name
}

// Fingerprint implements Type interface.
func (bt BaseType) Fingerprint() string {
	return fmt.Sprintf("base: %s", bt.name)
}

func TypeDecimal(precision, scale int) BaseType {
	switch {
	case precision == 0:
		return BaseType{name: "DECIMAL()"}
	case precision != 0 && scale == 0:
		return BaseType{name: fmt.Sprintf("DECIMAL(%d)", precision)}
	default:
		return BaseType{name: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)}
	}
}

func TypeReal() BaseType {
	return BaseType{name: "REAL"}
}

func TypeSerial() BaseType {
	return BaseType{name: "SERIAL"}
}

func TypeSerialSmall() BaseType {
	return BaseType{name: "SMALLSERIAL"}
}

func TypeSerialBig() BaseType {
	return BaseType{name: "BIGSERIAL"}
}

func TypeInteger() BaseType {
	return BaseType{name: "INTEGER"}
}

func TypeIntegerSmall() BaseType {
	return BaseType{name: "SMALLINT"}
}

func TypeIntegerBig() BaseType {
	return BaseType{name: "BIGINT"}
}

func TypeIntegerArray(l int64) BaseType {
	if l == 0 {
		return BaseType{name: "INTEGER[]"}
	}
	return BaseType{name: fmt.Sprintf("INTEGER[%d]", l)}
}

func TypeNumeric(precision, scale int) BaseType {
	switch {
	case precision == 0:
		return BaseType{name: "NUMERIC()"}
	case precision != 0 && scale == 0:
		return BaseType{name: fmt.Sprintf("NUMERIC(%d)", precision)}
	default:
		return BaseType{name: fmt.Sprintf("NUMERIC(%d,%d)", precision, scale)}
	}
}

func TypeDoublePrecision() BaseType {
	return BaseType{name: "DOUBLE PRECISION"}
}

func TypeBool() BaseType {
	return BaseType{name: "BOOL"}
}

func TypeText() BaseType {
	return BaseType{name: "TEXT"}
}

func TypeTextArray(l int64) BaseType {
	if l == 0 {
		return BaseType{name: "TEXT[]"}
	}
	return BaseType{name: fmt.Sprintf("TEXT[%d]", l)}
}

func TypeVarchar(l int64) BaseType {
	return BaseType{name: fmt.Sprintf("VARCHAR(%d)", l)}
}

func TypeBytea() BaseType {
	return BaseType{name: "BYTEA"}
}

func TypeTimestamp() BaseType {
	return BaseType{name: "TIMESTAMP"}
}

func TypeTimestampTZ() BaseType {
	return BaseType{name: "TIMESTAMPTZ"}
}

type CompositeType struct {
	name       string
	Attributes []*Attribute
}

// SQL implements Stringer interface.
func (ct CompositeType) String() string {
	return ""
}

// Fingerprint implements Type interface.
func (ct CompositeType) Fingerprint() string {
	return fmt.Sprintf("composite: %v", ct)
}

func TypeComposite(name string, attributes ...*Attribute) CompositeType {
	return CompositeType{
		name:       name,
		Attributes: attributes,
	}
}

type EnumeratedType struct {
	name  string
	Enums []string
}

// SQL implements Stringer interface.
func (et EnumeratedType) String() string {
	return et.name
}

// Fingerprint implements Type interface.
func (et EnumeratedType) Fingerprint() string {
	return fmt.Sprintf("enumarated: %v", et)
}

func TypeEnumerated(name string, enums ...string) EnumeratedType {
	return EnumeratedType{
		name:  name,
		Enums: enums,
	}
}

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

func TypePseudo(name string) PseudoType {
	return PseudoType{
		name: name,
	}
}

type MappableType struct {
	from    Type
	Mapping []Type
}

// String implements Stringer interface.
func (mt MappableType) String() string {
	return mt.from.String()
}

// Fingerprint implements Type interface.
func (mt MappableType) Fingerprint() string {
	return fmt.Sprintf("mappable: %v", mt)
}

func TypeMappable(from Type, mapping ...Type) MappableType {
	return MappableType{
		from:    from,
		Mapping: mapping,
	}
}
