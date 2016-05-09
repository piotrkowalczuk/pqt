package pqtgo

// Arguemnts ...
type Arguments struct {
	args []interface{}
}

// NewArguments allocates new Arguments object with given size.
func NewArguments(l int) *Arguments {
	return &Arguments{args: make([]interface{}, 0, l)}
}

// Add appends list with new element.
func (a *Arguments) Add(arg interface{}) {
	a.args = append(a.args, arg)
}

// Len returns number of arguments.
func (a *Arguments) Len() int64 {
	return int64(len(a.args))
}

// Slice returns all arguments stored as a slice.
func (a *Arguments) Slice() []interface{} {
	return a.args
}
