package pqtgo

import (
	"fmt"
	"go/types"
	"reflect"
)

// BuiltinType ...
type BuiltinType types.BasicKind

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

// Fingerprint implements Type interface.
func (bt BuiltinType) Fingerprint() string {
	return fmt.Sprintf("gobuiltin: %v", bt)
}

// CustomType ...
type CustomType struct {
	src                                             interface{}
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

// TypeCustom ...
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

// ValueOf ...
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

// TypeOf ...
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
