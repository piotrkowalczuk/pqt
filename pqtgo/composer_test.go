package pqtgo

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func ExampleComposer_Read() {
	com := NewComposer(0)
	buf := bytes.NewBufferString("SELECT * FROM user")
	arg := 1

	// lets imagine that not always argument is present
	if arg > 0 {
		_, _ = com.WriteString("age = ")
		_ = com.WritePlaceholder()
		com.Add(arg)
		com.Dirty = true // something was written, lets mark composer as dirty
	}

	if com.Dirty {
		buf.WriteString(" WHERE ")
		buf.ReadFrom(com)
	}

	fmt.Println(strings.TrimSpace(
		buf.String(),
	))
	fmt.Print(com.Args())
	// Output:
	// SELECT * FROM user WHERE age = $1
	// [1]
}

var (
	benchPlaceholderString string
)

func BenchmarkComposer_WritePlaceholder(b *testing.B) {
	com := NewComposer(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		com.WritePlaceholder()
	}

	benchPlaceholderString = com.String()
}

func TestComposer_WritePlaceholder(t *testing.T) {
	expected := "$1$2$3"
	com := NewComposer(0)
	com.WritePlaceholder()
	com.WritePlaceholder()
	com.WritePlaceholder()

	if com.String() != expected {
		t.Errorf("unexpected buffer output, expeted %s but got %s", expected, com.String())
	}
}

func TestComposer(t *testing.T) {
	com := NewComposer(0)
	expected := 100

	for i := 1; i < expected; i++ {
		if err := com.WritePlaceholder(); err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
	}

	if com.counter != expected {
		t.Errorf("count value, expected %d got %d", expected, com.counter)
	}
}
