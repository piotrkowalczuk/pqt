package testutil

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	"github.com/piotrkowalczuk/pqt/internal/print"
)

var Simple bool

func AssertOutput(t *testing.T, p print.Printer, e string) {
	t.Helper()

	got, err := format.Source(p.Bytes())
	if err != nil {
		t.Fatalf("unexpected printer formatting error: %s\n\n%s", err.Error(), p.String())
	}
	exp, err := format.Source([]byte(e))
	if err != nil {
		t.Fatalf("unexpected formatting error: %s", err.Error())
	}
	AssertGoCode(t, string(exp), string(got))
}

func AssertGoCode(t *testing.T, s1, s2 string) {
	t.Helper()

	tmp1 := strings.Split(s1, "\n")
	tmp2 := strings.Split(s2, "\n")
	if s1 != s2 {
		if Simple {
			t.Errorf("wrong output, expected:\n'%s'\nbut got:\n'%s'", s1, s2)
			return
		}
		b := bytes.NewBuffer(nil)
		for i, diff := range difflib.Diff(tmp1, tmp2) {
			p := strings.Replace(diff.Payload, "\t", "\\t", -1)
			switch diff.Delta {
			case difflib.Common:
				if testing.Verbose() {
					fmt.Fprintf(b, "%20d %s %s\n", i, diff.Delta.String(), p)
				}
			case difflib.LeftOnly:
				fmt.Fprintf(b, "\033[31m%20d %s %s\033[39m\n", i, diff.Delta.String(), p)
			case difflib.RightOnly:
				fmt.Fprintf(b, "\033[32m%20d %s %s\033[39m\n", i, diff.Delta.String(), p)
			}
		}
		t.Errorf(b.String())
	}
}
