package print

import (
	"bytes"
	"fmt"
)

type Printer struct {
	bytes.Buffer
	Err error
}

func (w *Printer) Print(args ...interface{}) {
	if w.Err != nil {
		return
	}
	if _, err := fmt.Fprint(w, args...); err != nil {
		w.Err = err
	}
}

func (w *Printer) Printf(format string, args ...interface{}) {
	if w.Err != nil {
		return
	}
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		w.Err = err
	}
}
func (w *Printer) Println(args ...interface{}) {
	if w.Err != nil {
		return
	}
	if _, err := fmt.Fprintln(w, args...); err != nil {
		w.Err = err
	}
}
