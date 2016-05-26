package pqt

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

const (
	arraySeparator     = ","
	arrayBeginningChar = "{"
	arrayEndChar       = "}"
)

// ArrayInt64 is a slice of int64s that implements necessary interfaces.
type ArrayInt64 []int64

// Scan satisfy sql.Scanner interface.
func (a *ArrayInt64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(ArrayInt64, 0)
		}
		return nil
	}

	var tmp []string
	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("pqt: expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf(`pqt: expected to get source argument in format "{1,2,...,N}", but got %s`, srcs)
	}

	if l == 2 {
		*a = make(ArrayInt64, 0)
		return nil
	}

	if string(srcs[0]) != arrayBeginningChar || string(srcs[l-1]) != arrayEndChar {
		return fmt.Errorf(`pqt: expected to get source argument in format "{1,2,...,N}", but got %s`, srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), arraySeparator)
	*a = make(ArrayInt64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf(`pqt: expected to get source argument in format "{1,2,...,N}", but got %s at index %d`, v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a ArrayInt64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(arrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(arraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatInt(v, 10)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(arrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ArrayString is a slice of strings that implements necessary interfaces.
type ArrayString []string

// Scan satisfy sql.Scanner interface.
func (a *ArrayString) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(ArrayString, 0)
		}
		return nil
	}

	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("pqt: expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf(`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got %s`, srcs)
	}

	if string(srcs[0]) != arrayBeginningChar || string(srcs[l-1]) != arrayEndChar {
		return fmt.Errorf(`pqt: expected to get source argument in format "{text1,text2,...,textN}", but got %s`, srcs)
	}

	*a = strings.Split(string(srcs[1:l-1]), arraySeparator)

	return nil
}

// Value satisfy driver.Valuer interface.
func (a ArrayString) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(arrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(arraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err = buffer.WriteString(v); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(arrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ArrayFloat64 is a slice of int64s that implements necessary interfaces.
type ArrayFloat64 []float64

// Scan satisfy sql.Scanner interface.
func (a *ArrayFloat64) Scan(src interface{}) error {
	if src == nil {
		if a == nil {
			*a = make(ArrayFloat64, 0)
		}
		return nil
	}

	var tmp []string
	var srcs string

	switch t := src.(type) {
	case []byte:
		srcs = string(t)
	case string:
		srcs = t
	default:
		return fmt.Errorf("pqt: expected slice of bytes or string as a source argument in Scan, not %T", src)
	}

	l := len(srcs)

	if l < 2 {
		return fmt.Errorf(`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got %s`, srcs)
	}

	if l == 2 {
		*a = make(ArrayFloat64, 0)
		return nil
	}

	if string(srcs[0]) != arrayBeginningChar || string(srcs[l-1]) != arrayEndChar {
		return fmt.Errorf(`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got %s`, srcs)
	}

	tmp = strings.Split(string(srcs[1:l-1]), arraySeparator)
	*a = make(ArrayFloat64, 0, len(tmp))
	for i, v := range tmp {
		j, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf(`pqt: expected to get source argument in format "{1.3,2.4,...,N.M}", but got %s at index %d`, v, i)
		}

		*a = append(*a, j)
	}

	return nil
}

// Value satisfy driver.Valuer interface.
func (a ArrayFloat64) Value() (driver.Value, error) {
	var (
		buffer bytes.Buffer
		err    error
	)

	if _, err = buffer.WriteString(arrayBeginningChar); err != nil {
		return nil, err
	}

	for i, v := range a {
		if i > 0 {
			if _, err := buffer.WriteString(arraySeparator); err != nil {
				return nil, err
			}
		}
		if _, err := buffer.WriteString(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
			return nil, err
		}
	}

	if _, err = buffer.WriteString(arrayEndChar); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
