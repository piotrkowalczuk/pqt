package pqtgo

import "bytes"

// Criterion ...
type Criterion interface {
	Criteria(*bytes.Buffer, *PlaceholderWriter, *Arguments, string) (int64, error)
}
