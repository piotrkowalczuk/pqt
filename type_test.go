package pqt_test

import (
	"testing"

	"reflect"

	"github.com/piotrkowalczuk/pqt"
)

func TestTypeSerialSmall(t *testing.T) {
	assertType(t, "SMALLSERIAL", pqt.TypeSerialSmall())
}

func TestTypeSerialBig(t *testing.T) {
	assertType(t, "BIGSERIAL", pqt.TypeSerialBig())
}

func TestTypeUUID(t *testing.T) {
	assertType(t, "UUID", pqt.TypeUUID())
}

func TestTypeCharacter(t *testing.T) {
	assertType(t, "CHARACTER[100]", pqt.TypeCharacter(100))
}

func TestTypeBytea(t *testing.T) {
	assertType(t, "BYTEA", pqt.TypeBytea())
}

func TestTypeTimestamp(t *testing.T) {
	assertType(t, "TIMESTAMP", pqt.TypeTimestamp())
}
func TestTypeTimestampTZ(t *testing.T) {
	assertType(t, "TIMESTAMPTZ", pqt.TypeTimestampTZ())
}
func TestTypeJSON(t *testing.T) {
	assertType(t, "JSON", pqt.TypeJSON())
}
func TestTypeJSONB(t *testing.T) {
	assertType(t, "JSONB", pqt.TypeJSONB())
}

func TestTypeIntegerSmall(t *testing.T) {
	assertType(t, "SMALLINT", pqt.TypeIntegerSmall())
}

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
	expected := "DOUBLE PRECISION[]"
	got := pqt.TypeDoubleArray(0)

	assertType(t, expected, got)
}

func TestTypeDoubleArray(t *testing.T) {
	expected := "DOUBLE PRECISION[100]"
	got := pqt.TypeDoubleArray(100)

	assertType(t, expected, got)
}

func TestTypeReal(t *testing.T) {
	expected := "REAL"
	got := pqt.TypeReal()
	assertType(t, expected, got)
}

func TestTypeDoublePrecision(t *testing.T) {
	expected := "DOUBLE PRECISION"
	got := pqt.TypeDoublePrecision()
	assertType(t, expected, got)
}

func TestTypeTextArray_zero(t *testing.T) {
	expected := "TEXT[]"
	got := pqt.TypeTextArray(0)
	assertType(t, expected, got)
}

func TestTypeTextArray(t *testing.T) {
	expected := "TEXT[100]"
	got := pqt.TypeTextArray(100)
	assertType(t, expected, got)
}

func TestTypeVarchar_zero(t *testing.T) {
	expected := "VARCHAR"
	got := pqt.TypeVarchar(0)
	assertType(t, expected, got)
}

func TestTypeVarchar(t *testing.T) {
	expected := "VARCHAR(100)"
	got := pqt.TypeVarchar(100)
	assertType(t, expected, got)
}

func assertType(t *testing.T, expected string, got pqt.Type) {
	if got.String() != expected {
		t.Errorf("unexpected sql representation, expected %s got %s", expected, got.String())
	}
}

func TestBaseType_Fingerprint(t *testing.T) {
	if pqt.TypeText().Fingerprint() != "base: TEXT" {
		t.Errorf("wrong fingerprint: %s", pqt.TypeText().Fingerprint())
	}
}

func TestTypeEnumerated(t *testing.T) {
	given := pqt.TypeEnumerated("pets", "cat", "dog", "pig")
	assertType(t, "pets", given)
	if !reflect.DeepEqual(given.Enums, []string{"cat", "dog", "pig"}) {
		t.Errorf("wrong set of enums: %v", given.Enums)
	}
	if given.Fingerprint() != "enumarated: pets" {
		t.Errorf("wrong fingerprint: %s", given.Fingerprint())
	}
}

func TestTypeMappable(t *testing.T) {
	given := pqt.TypeMappable(pqt.TypeInteger(), pqt.TypeText())
	assertType(t, "INTEGER", given)
	if given.Fingerprint() != "mappable: INTEGER" {
		t.Errorf("wrong fingerprint: %s", given.Fingerprint())
	}
}
