package pqtgo_test

import (
	"go/types"
	"testing"

	"reflect"

	"github.com/piotrkowalczuk/pqt/pqtgo"
)

func TestBuiltinType_String(t *testing.T) {
	cases := map[string]types.BasicKind{
		"invalid":    types.Invalid,
		"bool":       types.Bool,
		"int":        types.Int,
		"int8":       types.Int8,
		"int16":      types.Int16,
		"int32":      types.Int32,
		"int64":      types.Int64,
		"uint":       types.Uint,
		"uint8":      types.Uint8,
		"uint16":     types.Uint16,
		"uint32":     types.Uint32,
		"uint64":     types.Uint64,
		"float32":    types.Float32,
		"float64":    types.Float64,
		"complex64":  types.Complex64,
		"complex128": types.Complex128,
		"string":     types.String,
	}
	for exp, kind := range cases {
		t.Run(exp, func(t *testing.T) {
			got := pqtgo.BuiltinType(kind).String()
			if got != exp {
				t.Errorf("wrong string representation of go builtin type: %s", got)
			}
			got = pqtgo.BuiltinType(kind).Fingerprint()
			if got != "gobuiltin: "+exp {
				t.Errorf("wrong fingerprint of go builtin type: %s", got)
			}
		})
	}
}

func TestTypeCustom(t *testing.T) {
	type mandatory struct {
		X string
	}
	type optional struct {
		X string
	}
	type criteria struct {
		X string
	}

	m := &mandatory{}
	o := &optional{}
	c := &criteria{}

	got := pqtgo.TypeCustom(m, o, c)
	valueOfM := got.ValueOf(pqtgo.ModeMandatory)
	valueOfO := got.ValueOf(pqtgo.ModeOptional)
	valueOfC := got.ValueOf(pqtgo.ModeCriteria)
	valueOfD := got.ValueOf(pqtgo.ModeDefault)

	if valueOfD != nil {
		t.Error("value of default should be nil")
	}
	if !reflect.DeepEqual(m, valueOfM) {
		t.Errorf("wrong mandatory value found: %v", valueOfM)
	}
	if !reflect.DeepEqual(o, valueOfO) {
		t.Errorf("wrong optional value found: %v", valueOfO)
	}
	if !reflect.DeepEqual(c, valueOfC) {
		t.Errorf("wrong criteria value found: %v", valueOfC)
	}

	typeOfM := got.TypeOf(pqtgo.ModeMandatory)
	typeOfO := got.TypeOf(pqtgo.ModeOptional)
	typeOfC := got.TypeOf(pqtgo.ModeCriteria)
	typeOfD := got.TypeOf(pqtgo.ModeDefault)

	if typeOfD != nil {
		t.Error("type of default should be nil")
	}
	if typeOfM.String() != "*pqtgo_test.mandatory" {
		t.Errorf("wrong mandatory type found: %s", typeOfM.String())
	}
	if typeOfO.String() != "*pqtgo_test.optional" {
		t.Errorf("wrong optional type found: %s", typeOfO.String())
	}
	if typeOfC.String() != "*pqtgo_test.criteria" {
		t.Errorf("wrong criteria type found: %s", typeOfC.String())
	}

	gotS := got.String()
	if gotS != "*pqtgo_test.mandatory/*pqtgo_test.optional/*pqtgo_test.criteria" {
		t.Errorf("wrong string representation: %s", got.String())
	}

	gotF := got.Fingerprint()
	if gotF != "gocustomtype: *pqtgo_test.mandatory/*pqtgo_test.optional/*pqtgo_test.criteria" {
		t.Errorf("wrong fingerprint: %s", got.Fingerprint())
	}
}
