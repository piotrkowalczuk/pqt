package pqtgo

import (
	"reflect"
	"testing"

	"github.com/piotrkowalczuk/qtypes"
)

func TestWriteCompositionQueryInt64(t *testing.T) {
	cases := map[string]struct {
		sel  string
		obj  *qtypes.Int64
		opt  *CompositionOpts
		exp  string
		args []interface{}
	}{
		"null": {
			sel:  "x",
			obj:  qtypes.NullInt64(),
			opt:  And,
			exp:  " AND x IS NULL",
			args: []interface{}{},
		},
		"equal": {
			sel:  "x",
			obj:  qtypes.EqualInt64(1),
			opt:  And,
			exp:  " AND x = $1",
			args: []interface{}{int64(1)},
		},
		"greater": {
			sel:  "age",
			obj:  qtypes.GreaterInt64(1),
			opt:  And,
			exp:  " AND age > $1",
			args: []interface{}{int64(1)},
		},
		"greater-equal": {
			sel:  "age",
			obj:  qtypes.GreaterEqualInt64(1),
			opt:  And,
			exp:  " AND age >= $1",
			args: []interface{}{int64(1)},
		},
		"less": {
			sel:  "age",
			obj:  qtypes.LessInt64(1),
			opt:  And,
			exp:  " AND age < $1",
			args: []interface{}{int64(1)},
		},
		"less-equal": {
			sel:  "age",
			obj:  qtypes.LessEqualInt64(1),
			opt:  And,
			exp:  " AND age <= $1",
			args: []interface{}{int64(1)},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			com := NewComposer(0)
			com.Dirty = true
			err := WriteCompositionQueryInt64(c.obj, c.sel, com, c.opt)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			got := com.String()
			if c.exp != got {
				t.Errorf("wrong query, expected '%s' but got '%s'", c.exp, got)
			}
			if !reflect.DeepEqual(c.args, com.Args()) {
				t.Errorf("wrong arguments, expected %v but got %v", c.args, com.Args())
			}
		})
	}
}

func TestWriteCompositionQueryString(t *testing.T) {
	cases := map[string]struct {
		sel  string
		obj  *qtypes.String
		opt  *CompositionOpts
		exp  string
		args []interface{}
	}{
		"equal": {
			sel:  "name",
			obj:  qtypes.EqualString("John"),
			opt:  And,
			exp:  " AND name = $1",
			args: []interface{}{"John"},
		},
		"has-prefix": {
			sel:  "name",
			obj:  qtypes.HasPrefixString("dr."),
			opt:  And,
			exp:  " AND name LIKE $1",
			args: []interface{}{string("dr.%")},
		},
		"has-suffix": {
			sel:  "name",
			obj:  qtypes.HasSuffixString("jr"),
			opt:  And,
			exp:  " AND name LIKE $1",
			args: []interface{}{string("%jr")},
		},
		"substring": {
			sel:  "name",
			obj:  qtypes.SubString("xyz"),
			opt:  And,
			exp:  " AND name LIKE $1",
			args: []interface{}{string("%xyz%")},
		},
		"in": {
			sel: "name",
			obj: &qtypes.String{
				Type:   qtypes.QueryType_IN,
				Values: []string{"a", "b", "c"},
				Valid:  true,
			},
			opt:  And,
			exp:  " AND name IN ($1,$2,$3)",
			args: []interface{}{"a", "b", "c"},
		},
		"null": {
			sel:  "name",
			obj:  qtypes.NullString(),
			opt:  And,
			exp:  " AND name IS NULL",
			args: []interface{}{},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			com := NewComposer(0)
			com.Dirty = true
			err := WriteCompositionQueryString(c.obj, c.sel, com, c.opt)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			got := com.String()
			if c.exp != got {
				t.Errorf("wrong query, expected '%s' but got '%s'", c.exp, got)
			}
			if !reflect.DeepEqual(c.args, com.Args()) {
				t.Errorf("wrong arguments, expected %v but got %v", c.args, com.Args())
			}
		})
	}
}
