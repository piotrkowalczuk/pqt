package pqt

type Function struct {
}

func FunctionNow() *Function {
	return &Function{}
}

func (f *Function) Name() string {
	return ""
}
