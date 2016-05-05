package pqt

import (
	"bytes"
	"strconv"
)

type PlaceholderWriter struct {
	counter int64
}

func NewPlaceholderWriter() *PlaceholderWriter {
	return &PlaceholderWriter{counter: 1}
}

// WriteTo implements io WriterTo interface.
func (pw *PlaceholderWriter) WriteTo(buf *bytes.Buffer) (int64, error) {
	if _, err := buf.WriteString("$"); err != nil {
		return 0, err
	}
	if _, err := buf.WriteString(strconv.FormatInt(pw.counter, 10)); err != nil {
		return 0, err
	}
	if _, err := buf.WriteString(" "); err != nil {
		return 0, err
	}

	pw.counter++
	return 0, nil
}

// Count ...
func (pw *PlaceholderWriter) Count() int64 {
	return pw.counter
}
