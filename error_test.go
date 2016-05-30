package pqt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func ExampleErrorConstraint() {
	expected := "pq: something goes wrong"
	err := &pq.Error{
		Table:      "user",
		Constraint: expected,
	}

	switch ErrorConstraint(err) {
	case expected:
		fmt.Println("expected constraint")
	default:
		fmt.Println("unknown constraint")
	}

	// Output:
	// expected constraint
}

func TestErrorConstraint(t *testing.T) {
	expected := "something"
	err := &pq.Error{
		Constraint: expected,
	}
	got := ErrorConstraint(err)
	if got != expected {
		t.Fatalf("wrong constraint, expected %s but got %s", expected, got)
	}
}

func TestErrorConstraint_nil(t *testing.T) {
	got := ErrorConstraint(nil)
	if got != "" {
		t.Fatalf("wrong constraint, expected empty string but got %s", got)
	}
}

func TestErrorConstraint_nonSQL(t *testing.T) {
	err := errors.New("normal error")
	got := ErrorConstraint(err)
	if got != "" {
		t.Fatalf("wrong constraint, expected empty string but got %s", got)
	}
}
