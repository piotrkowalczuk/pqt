package gogen_test

import (
	"flag"
	"os"
	"testing"

	"github.com/piotrkowalczuk/pqt/internal/testutil"
)

func TestMain(m *testing.M) {
	flag.BoolVar(&testutil.Simple, "simple", false, "if true, tests will print out in simplified form")
	flag.Parse()

	os.Exit(m.Run())
}

type testColumn struct {
	name, kind string
}
