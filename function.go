package pqt

const (
	// FunctionBehaviourVolatile indicates that the function value can change even within a single table scan,
	// so no optimizations can be made.
	// Relatively few database functions are volatile in this sense; some examples are random(), currval(), timeofday().
	// But note that any function that has side-effects must be classified volatile, even if its result is quite predictable,
	// to prevent calls from being optimized away; an example is setval().
	FunctionBehaviourVolatile FunctionBehaviour = iota
	// FunctionBehaviourImmutable indicates that the function cannot modify the database and always returns the same result when given the same argument values;
	// that is, it does not do database lookups or otherwise use information not directly present in its argument list.
	// If this option is given, any call of the function with all-constant arguments can be immediately replaced with the function value.
	FunctionBehaviourImmutable
	// FunctionBehaviourStable indicates that the function cannot modify the database,
	// and that within a single table scan it will consistently return the same result for the same argument values,
	// but that its result could change across SQL statements.
	// This is the appropriate selection for functions whose results depend on database lookups,
	// parameter variables (such as the current time zone), etc.
	// (It is inappropriate for AFTER triggers that wish to query rows modified by the current command.)
	// Also note that the current_timestamp family of functions qualify as stable, since their values do not change within a transaction.
	FunctionBehaviourStable
)

type FunctionBehaviour int

// Function ...
type Function struct {
	Name      string
	BuiltIn   bool
	Type      Type
	Body      string
	Behaviour FunctionBehaviour
	Args      []*FunctionArg
}

// FunctionArg ...
type FunctionArg struct {
	Name string
	Type Type
}

// FunctionNow ...
func FunctionNow() *Function {
	return &Function{
		Name:    "now",
		BuiltIn: true,
		Type:    TypeTimestampTZ(),
	}
}
