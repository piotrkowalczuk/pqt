package pqt

import (
	"errors"
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
