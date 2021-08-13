package parser

type ValueType int

const (
	ValueTypeString ValueType = iota
	ValueTypeNumber
	ValueTypeBool
)

type FlagValues map[string]Value

func (fv FlagValues) Get(name string) (Value, bool) {
	if v, ok := fv[name]; ok {
		return v, ok
	}

	return NullValue, false
}

func (fv FlagValues) Set(name string, value Value) {
	fv[name] = value
}

type Flags map[string]*Flag

func (f Flags) Get(name string) *Flag {
	if flag, ok := f[name]; ok {
		return flag
	}

	return nil
}

func (f Flags) Set(flag *Flag) {
	f[flag.Name] = flag
}

type Options map[string]bool

func (o Options) Get(name string) bool {
	return o[name]
}

func (o Options) Set(name string) {
	o[name] = true
}

type CommandType string

const (
	CommandTypeUser   CommandType = "user"
	CommandTypeSystem CommandType = "system"
)

type TestExec interface {
	Exec(ctx Context, flags FlagValues, options Options) (Value, error)
}

type ExecFunc func(Context, FlagValues, Options) (Value, error)
type SystemExecFunc func(SystemContext, Flags, Options) (Value, error)
type OutFunc func(SystemContext, Flags, Options)

type Command struct {
	Type           CommandType
	Path           []string
	Name           string
	ExecFunc       ExecFunc
	SystemExecFunc SystemExecFunc
	OutFunc        OutFunc
	Flags          Flags
	Options        Options
	MandatoryFlags []string

	unnamedFlags map[uint]*Flag
}

func (c *Command) Exec(ctx SystemContext, flags Flags) (Value, error) {
	for _, f := range c.MandatoryFlags {
		if flags.Get(f) == nil {
			return nil, ErrNoMandatoryFlag
		}
	}

	if c.Type == CommandTypeSystem && c.SystemExecFunc != nil {
		return c.SystemExecFunc(ctx, flags, c.Options)
	}

	if c.Type == CommandTypeUser && c.ExecFunc != nil {
		flagValues := make(FlagValues)
		for _, flag := range flags {
			flagValues.Set(flag.Name, flag.Value(ctx))
		}

		return c.ExecFunc(ctx, flagValues, c.Options)
	}

	return nil, nil
}

func (c *Command) Out(ctx SystemContext, inFlags Flags) {
	if c.Type == CommandTypeSystem && c.OutFunc != nil {
		c.OutFunc(ctx, inFlags, c.Options)
	}
}

func (c *Command) Copy() *Command {
	copied := new(Command)
	copied.Type = c.Type
	copied.Name = c.Name
	copied.Path = append(copied.Path, c.Path...)
	copied.SystemExecFunc = c.SystemExecFunc
	copied.ExecFunc = c.ExecFunc
	copied.OutFunc = c.OutFunc
	copied.Flags = make(map[string]*Flag)
	copied.Options = make(map[string]bool)

	copied.unnamedFlags = make(map[uint]*Flag, len(c.unnamedFlags))
	for i, flag := range c.unnamedFlags {
		copied.unnamedFlags[i] = flag
	}

	copied.MandatoryFlags = make([]string, len(c.MandatoryFlags))
	copy(copied.MandatoryFlags, c.MandatoryFlags)

	return copied
}

func (c *Command) UnnamedFlag(number uint) *Flag {
	return c.unnamedFlags[number]
}

type Result struct {
	Message string
}
