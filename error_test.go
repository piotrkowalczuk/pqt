package pqt

import (
	"testing"

	"github.com/lib/pq"
)

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
