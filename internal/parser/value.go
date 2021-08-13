package parser

type Valuer interface {
	Value(ctx SystemContext) Value
}

type Value interface {
	Bool() bool
	Number() int
	String() string

	IsBool() bool
	IsString() bool
	IsNumber() bool

	Equal(v Value) bool
	Less(v Value) bool
	Greater(v Value) bool
}

var (
	NullValue = NewStringValue("")
)
