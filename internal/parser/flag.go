package parser

type Flag struct {
	Name      string
	Mandatory bool
	Number    uint
	ValueType
	expression Expression
}

func (f *Flag) Set(e Expression) {
	f.expression = e
}

func (f *Flag) Expression() Expression {
	return f.expression
}

func (f *Flag) Value(ctx SystemContext) Value {
	return f.expression.Value(ctx)
}

func (f *Flag) Copy() *Flag {
	copied := new(Flag)
	copied.Name = f.Name
	copied.Mandatory = f.Mandatory
	copied.ValueType = f.ValueType

	return copied
}

type Option struct {
	Name string
}

func (o *Option) Copy() *Option {
	copied := new(Option)
	copied.Name = o.Name

	return copied
}
