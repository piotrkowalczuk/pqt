package pqt

import (
	"bytes"
	"testing"
)

var (
	benchPlaceholderString string
)

func BenchmarkPlaceholderWriter_WriteTo(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	pw := NewPlaceholderWriter()

	for n := 0; n < b.N; n++ {
		pw.WriteTo(buf)
	}

	benchPlaceholderString = buf.String()
}

func TestPlaceholderWriter_WriteTo(t *testing.T) {
	expected := "$1 $2 $3 "
	buf := bytes.NewBuffer(nil)
	pw := NewPlaceholderWriter()
	pw.WriteTo(buf)
	pw.WriteTo(buf)
	pw.WriteTo(buf)

	if buf.String() != expected {
		t.Errorf("unexpected buffer output, expeted %s but got %s", expected, buf.String())
	}
}

func TestPlaceholderWriter_Count(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	pw := NewPlaceholderWriter()
	expected := 100

	for i := 1; i < expected; i++ {
		if _, err := pw.WriteTo(buf); err != nil {
			t.Fatal("unexpected error: %s", err.Error())
		}
	}

	if pw.Count() != int64(expected) {
		t.Errorf("count value, expected %d got %d", expected, pw.Count())
	}
}
