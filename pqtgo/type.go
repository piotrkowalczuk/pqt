package pqtgo

import (
	"fmt"
	"go/types"
	"reflect"
)

const (
	// ModeDefault is a default mode.
	// Modes allows to express context in what column/property is used in generated Go code.
	// Main purpose of it is to define clear contract in what contexts each column/property can be generated.
	ModeDefault = iota
	// ModeMandatory is used when property is mandatory.
	// It could be the case for insert statements when property corresponding column is not nullable.
	ModeMandatory
	// ModeOptional is mode used when property is optional in given context or in general.
	// Example:
	// 	Insert statement of optional property.
	// 	Partial update statement of mandatory property.
	ModeOptional
	// ModeCriteria indicates that property is used in context of querying.
	// For example during FindIter generation.
	ModeCriteria
)

// BuiltinType is simple alias for types.BasicKind.
type BuiltinType types.BasicKind

// String implements pqt Type interface.
func (bt BuiltinType) String() string {
	switch types.BasicKind(bt) {
	case types.Bool:
		return "bool"
	case types.Int:
		return "int"
	case types.Int8:
		return "int8"
	case types.Int16:
		return "int16"
	case types.Int32:
		return "int32"
	case types.Int64:
		return "int64"
	case types.Uint:
		return "uint"
	case types.Uint8:
		return "uint8"
	case types.Uint16:
		return "uint16"
	case types.Uint32:
		return "uint32"
	case types.Uint64:
		return "uint64"
	case types.Float32:
		return "float32"
	case types.Float64:
		return "float64"
	case types.Complex64:
		return "complex64"
	case types.Complex128:
		return "complex128"
	case types.String:
		return "string"
	default:
		return "invalid"
	}
}

// Fingerprint implements pqt Type interface.
func (bt BuiltinType) Fingerprint() string {
	return fmt.Sprintf("gobuiltin: %v", bt)
}

// CustomType allows to create custom types from already existing types.
type CustomType struct {
	mandatory, optional, criteria                   interface{}
	mandatoryTypeOf, optionalTypeOf, criteriaTypeOf reflect.Type
}

// String implements Stringer interface.
func (ct CustomType) String() string {
	return fmt.Sprintf("%s/%s/%s", ct.mandatoryTypeOf.String(), ct.optionalTypeOf.String(), ct.criteriaTypeOf.String())
}

// Fingerprint implements Type interface.
func (ct CustomType) Fingerprint() string {
	return fmt.Sprintf("gocustomtype: %v", ct)
}

// TypeCustom allocates new CustomType using given arguments for each context: mandatory, optional and criteria.
func TypeCustom(m, o, c interface{}) CustomType {
	var mandatoryTypeOf, optionalTypeOf, criteriaTypeOf reflect.Type
	if m != nil {
		mandatoryTypeOf = reflect.ValueOf(m).Type()
	}
	if o != nil {
		optionalTypeOf = reflect.ValueOf(o).Type()
	}
	if c != nil {
		criteriaTypeOf = reflect.ValueOf(c).Type()
	}

	return CustomType{
		mandatory:       m,
		criteria:        c,
		optional:        o,
		mandatoryTypeOf: mandatoryTypeOf,
		optionalTypeOf:  optionalTypeOf,
		criteriaTypeOf:  criteriaTypeOf,
	}
}

// ValueOf returns type for given mode.
func (ct CustomType) ValueOf(m int32) interface{} {
	switch m {
	case ModeMandatory:
		return ct.mandatory
	case ModeOptional:
		return ct.optional
	case ModeCriteria:
		return ct.criteria
	default:
		return nil
	}
}

// TypeOf returns Go type of underlying pqt Type for given mode.
func (ct CustomType) TypeOf(m int32) reflect.Type {
	switch m {
	case ModeMandatory:
		return ct.mandatoryTypeOf
	case ModeOptional:
		return ct.optionalTypeOf
	case ModeCriteria:
		return ct.criteriaTypeOf
	default:
		return nil
	}
}
