package pqt_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestTypeIntegerSmallArray_zero(t *testing.T) {
	expected := "SMALLINT[]"
	got := pqt.TypeIntegerSmallArray(0)

	assertType(t, expected, got)
}

func TestTypeIntegerSmallArray(t *testing.T) {
	expected := "SMALLINT[100]"
	got := pqt.TypeIntegerSmallArray(100)

	assertType(t, expected, got)
}

func TestTypeIntegerArray_zero(t *testing.T) {
	expected := "INTEGER[]"
	got := pqt.TypeIntegerArray(0)

	assertType(t, expected, got)
}

func TestTypeIntegerArray(t *testing.T) {
	expected := "INTEGER[100]"
	got := pqt.TypeIntegerArray(100)

	assertType(t, expected, got)
}

func TestTypeIntegerBigArray_zero(t *testing.T) {
	expected := "BIGINT[]"
	got := pqt.TypeIntegerBigArray(0)

	assertType(t, expected, got)
}

func TestTypeIntegerBigArray(t *testing.T) {
	expected := "BIGINT[100]"
	got := pqt.TypeIntegerBigArray(100)

	assertType(t, expected, got)
}

func TestTypeDecimal_zeroPrecisionZeroScale(t *testing.T) {
	expected := "DECIMAL"
	got := pqt.TypeDecimal(0, 0)

	assertType(t, expected, got)
}

func TestTypeDecimal_zeroScale(t *testing.T) {
	expected := "DECIMAL(100)"
	got := pqt.TypeDecimal(100, 0)

	assertType(t, expected, got)
}

func TestTypeDecimal(t *testing.T) {
	expected := "DECIMAL(100,2)"
	got := pqt.TypeDecimal(100, 2)

	assertType(t, expected, got)
}

func TestTypeNumeric_zeroPrecisionZeroScale(t *testing.T) {
	expected := "NUMERIC"
	got := pqt.TypeNumeric(0, 0)

	assertType(t, expected, got)
}

func TestTypeNumeric_zeroScale(t *testing.T) {
	expected := "NUMERIC(100)"
	got := pqt.TypeNumeric(100, 0)

	assertType(t, expected, got)
}

func TestTypeNumeric(t *testing.T) {
	expected := "NUMERIC(100,2)"
	got := pqt.TypeNumeric(100, 2)

	assertType(t, expected, got)
}

func TestTypeDoubleArray_zero(t *testing.T) {
	expected := "DOUBLE[]"
	got := pqt.TypeDoubleArray(0)

	assertType(t, expected, got)
}

func TestTypeDoubleArray(t *testing.T) {
	expected := "DOUBLE[100]"
	got := pqt.TypeDoubleArray(100)

	assertType(t, expected, got)
}

func assertType(t *testing.T, expected string, got pqt.Type) {
	if got.String() != expected {
		t.Errorf("unexpected sql representation, expected %s got %s", expected, got.String())
	}
}
