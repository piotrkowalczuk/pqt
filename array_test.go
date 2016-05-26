package pqt_test

import (
	"testing"
	"time"

	"reflect"

	"github.com/piotrkowalczuk/pqt"
)

func TestArrayInt64_Value(t *testing.T) {
	success := map[string]pqt.ArrayInt64{
		"{1,2,3,4}": pqt.ArrayInt64{0: 1, 1: 2, 2: 3, 3: 4},
		"{}":        pqt.ArrayInt64{},
	}

SuccessLoop:
	for expected, array := range success {
		got, err := array.Value()
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}

		gots, ok := got.([]byte)
		if !ok {
			t.Errorf("wrong output type, expected slice of bytes, got %T", got)
			continue SuccessLoop
		}

		if expected != string(gots) {
			t.Errorf("wrong output, expected %s but got %s", expected, gots)
		}
	}
}

func TestArrayInt64_Scan(t *testing.T) {
	success := map[string]pqt.ArrayInt64{
		"{1,2,3,4}": pqt.ArrayInt64{0: 1, 1: 2, 2: 3, 3: 4},
		"{}":        pqt.ArrayInt64{},
	}

SuccessLoop:
	for src, expected := range success {
		var got pqt.ArrayInt64

		err := got.Scan(src)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("wrong output, expected %s but got %s", expected, got)
		}
	}

	fail := map[string]interface{}{
		`pqt: expected to get source argument in format "{1,2,...,N}", but got string1 at index 0`: "{string1,string2}",
		`pqt: expected to get source argument in format "{1,2,...,N}", but got }`:                  "}",
		`pqt: expected to get source argument in format "{1,2,...,N}", but got {`:                  "{",
		`pqt: expected to get source argument in format "{1,2,...,N}", but got 12412s at index 0`:  "{12412s}",
		`pqt: expected to get source argument in format "{1,2,...,N}", but got [1,2,3]`:            "[1,2,3]",
		`pqt: expected to get source argument in format "{1,2,...,N}", but got (1,2,3)`:            "(1,2,3)",
	}

FailLoop:
	for expected, src := range fail {
		var got pqt.ArrayInt64

		err := got.Scan(src)
		if err == nil {
			t.Errorf("expected error, got nil")
			continue FailLoop
		}
		if expected != err.Error() {
			t.Errorf("undexpected error, got:\n	%s\nbut expected:\n	%s", err.Error(), expected)
		}
	}
}

func TestArrayInt64_Scan_nil(t *testing.T) {
	var got pqt.ArrayInt64

	if err := got.Scan(nil); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if got != nil {
		t.Errorf("unexpected output, expected: %T\n	%s\n	but got: %T\n	%s,", nil, nil, got, got)

	}
}

func TestArrayString_Value(t *testing.T) {
	success := map[string]pqt.ArrayString{
		"{1,2,3,4}":             pqt.ArrayString{0: "1", 1: "2", 2: "3", 3: "4"},
		"{hehe1,string,some,'}": pqt.ArrayString{0: "hehe1", 1: "string", 2: "some", 3: "'"},
		"{}": pqt.ArrayString{},
	}

SuccessLoop:
	for expected, array := range success {
		got, err := array.Value()
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}

		gots, ok := got.([]byte)
		if !ok {
			t.Errorf("wrong output type, expected slice of bytes, got %T", got)
			continue SuccessLoop
		}

		if expected != string(gots) {
			t.Errorf("wrong output, expected %s but got %s", expected, gots)
		}
	}
}

func TestArrayString_Scan(t *testing.T) {
	success := map[string]pqt.ArrayString{
		"{1,2,3,4}": pqt.ArrayString{0: "1", 1: "2", 2: "3", 3: "4"},
	}

SuccessLoop:
	for src, expected := range success {
		var got pqt.ArrayString
		if err := got.Scan(src); err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}
		if !reflect.DeepEqual(expected, got) {
			t.Errorf("unexpected output, expected: %T\n	%s\n	but got: %T\n	%s,", expected, expected, got, got)

		}
	}

	fail := map[string]interface{}{
		`pqt: expected slice of bytes or string as a source argument in Scan, not int64`:                        int64(1),
		`pqt: expected slice of bytes or string as a source argument in Scan, not bool`:                         false,
		`pqt: expected slice of bytes or string as a source argument in Scan, not float64`:                      float64(12.2),
		`pqt: expected slice of bytes or string as a source argument in Scan, not time.Time`:                    time.Now(),
		`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got {`:                   "{",
		`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got }`:                   "}",
		`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got [text1,text2,text3]`: "[text1,text2,text3]",
		`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got (text1,text2,text3)`: "(text1,text2,text3)",
	}

FailLoop:
	for expected, src := range fail {
		var got pqt.ArrayString

		err := got.Scan(src)
		if err == nil {
			t.Errorf("expected error: %s", expected)
			continue FailLoop
		}
		if expected != err.Error() {
			t.Errorf("undexpected error, got:\n	%s\nbut expected:\n	%s", err.Error(), expected)
		}
	}
}

func TestArrayString_Scan_nil(t *testing.T) {
	var got, expected pqt.ArrayString
	err := got.Scan(nil)

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("unexpected output, expected: %T\n	%s\n	but got: %T\n	%s,", expected, expected, got, got)

	}
}

func TestArrayFloat64_Value(t *testing.T) {
	success := map[string]pqt.ArrayFloat64{
		"{1.1,2.2,3.5,4.65}": pqt.ArrayFloat64{0: 1.1, 1: 2.2, 2: 3.5, 3: 4.65},
		"{}":       				 pqt.ArrayFloat64{},
	}

SuccessLoop:
	for expected, array := range success {
		got, err := array.Value()
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}

		gots, ok := got.([]byte)
		if !ok {
			t.Errorf("wrong output type, expected slice of bytes, got %T", got)
			continue SuccessLoop
		}

		if expected != string(gots) {
			t.Errorf("wrong output, expected %s but got %s", expected, gots)
		}
	}
}

func TestArrayFloat64_Scan(t *testing.T) {
	success := map[string]pqt.ArrayFloat64{
		"{1.1,2.2,3.5,4.65}": pqt.ArrayFloat64{0: 1.1, 1: 2.2, 2: 3.5, 3: 4.65},
		"{}":       					pqt.ArrayFloat64{},
	}

SuccessLoop:
	for src, expected := range success {
		var got pqt.ArrayFloat64

		err := got.Scan(src)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			continue SuccessLoop
		}

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("wrong output, expected %s but got %s", expected, got)
		}
	}

	fail := map[string]interface{}{
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got string1 at index 0`: 			"{string1,string2}",
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got }`:                  			"}",
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got {`:                  			"{",
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got 12412s at index 0`:  			"{12412s}",
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got [1.1,2.2,3.5]`:            "[1.1,2.2,3.5]",
		`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got (1.1,2.2,3.5)`:            "(1.1,2.2,3.5)",
	}

FailLoop:
	for expected, src := range fail {
		var got pqt.ArrayFloat64

		err := got.Scan(src)
		if err == nil {
			t.Errorf("expected error, got nil")
			continue FailLoop
		}
		if expected != err.Error() {
			t.Errorf("undexpected error, got:\n	%s\nbut expected:\n	%s", err.Error(), expected)
		}
	}
}

func TestArrayFloat64_Scan_nil(t *testing.T) {
	var got pqt.ArrayFloat64

	if err := got.Scan(nil); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if got != nil {
		t.Errorf("unexpected output, expected: %T\n	%s\n	but got: %T\n	%s,", nil, nil, got, got)

	}
}
