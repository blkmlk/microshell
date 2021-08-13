package parser

import (
	"errors"
	"fmt"

	"github.com/blkmlk/microshell/internal/models"
)

var (
	ErrWrongRune    = errors.New("wrong rune")
	ErrWrongPayload = errors.New("wrong payload")
	ErrPanic        = errors.New("panic")
	ErrNotFinished  = errors.New("not finished")
)

type ResponseType string

type StateCommand int

const (
	StateCommandStart StateCommand = iota
	StateCommandPath
	StateCommandCommand
	StateCommandArgument
	StateCommandFlag
	StateFlagEqual
	StateFlagValue
	StateCommandOption
)

type commandExpression struct {
	relativeRoot *CommandTree
	flagTree     *CommandTree
	iterator     *commandIterator

	setRelativeRoot bool
	state           StateCommand
	prevRune        models.Rune
	started         bool

	// path + command
	currentCommand *Command

	// flag
	currentFlag         *Flag
	unnamedFlagPosition uint
	unnamedFlagValue    []models.Rune
	flags               Flags
	flagUsed            bool

	// brackets
	opened      int
	quoteOpened int
}

func NewCommandExpression(ctx SystemContext) Expression {
	return &commandExpression{
		relativeRoot: ctx.CommandRoot(),
		iterator:     ctx.CommandRoot().GetIterator(),
		state:        StateCommandStart,
		flags:        make(Flags),
	}
}

func (c *commandExpression) Type() ExpressionType {
	return ExpressionTypeCmd
}

func (c *commandExpression) Value(ctx SystemContext) Value {
	switch c.state {
	case StateCommandStart, StateCommandPath:
		return NullValue
	default:
		if c.currentCommand == nil {
			return NullValue
		}

		value, err := c.currentCommand.Exec(ctx, c.flags)

		if err != nil {
			// TODO: handle err
			return nil
		}

		if value == nil {
			value = NullValue
		}

		return value
	}
}

func (c *commandExpression) Complete(ctx SystemContext) *CompleteResponse {
	if c.iterator == nil {
		return nil
	}

	if c.state == StateFlagEqual {
		return nil
	}

	var resp CompleteResponse

	opts := c.iterator.NextOptions()

	for _, o := range opts.Options {
		resp.Options = append(resp.Options, &CompleteOption{Level: o.Level, Option: o.Name})
	}

	resp.Merged = opts.Merged

	return &resp
}

func (c *commandExpression) Add(ctx SystemContext, r models.Rune) *Response {
	var resp *Response

	switch {
	case r.Is('/'):
		resp = c.handleSlash(ctx)
	case r.Is(':'):
		resp = c.handleColon(ctx)
	case r.Is(' '):
		resp = c.handleSpace(ctx)
	case r.Is(';'):
		resp = c.handleSemicolon(ctx)
	case r.Is('='):
		resp = c.handleEqual()
	case r.IsLowerAlpha():
		resp = c.handleLowerAlpha(ctx, r)
	case r.IsNumber():
		resp = c.handleNumber(ctx)
	case r.Is('"'):
		resp = c.handleQuote(ctx)
	case r.Is('$'):
		resp = c.handleVariable(ctx)
	case r.Is('[') || r.Is(']') || r.Is('{') || r.Is('}'):
		resp = c.handleCommandList(ctx, r)
	case r.Is('(') || r.Is(')'):
		resp = c.handleMath(ctx, r)
	default:
		return NewResponse().WithError(ErrWrongRune)
	}

	if !((resp.Action() == ResponseRepeat || resp.Action() == ResponseGoOut) && r.IsSpace()) {
		c.prevRune = r
	}

	return resp
}

func (c *commandExpression) Close(ctx SystemContext) *CloseResponse {
	var resp CloseResponse

	switch c.state {
	case StateCommandStart:
		if !c.setRelativeRoot {
			ctx.SetCommandRoot(ctx.CommandTree())
		}
	case StateCommandPath:
		if c.setRelativeRoot {
			ctx.SetCommandRoot(c.relativeRoot)
		} else {
			ctx.SetCommandRoot(c.iterator.NextTree())
		}
	case StateCommandCommand:
		if c.currentCommand == nil {
			if !c.iterator.GoToEnd() {
				resp.Error = ErrWrongPayload
				return &resp
			}

			cmd, ok := c.iterator.Payload().(*Command)
			if !ok {
				resp.Error = ErrWrongPayload
				return &resp
			}
			if cmd != nil {
				c.currentCommand = cmd.Copy()
			}
		}
	case StateFlagValue:
		c.flags.Set(c.currentFlag)
	case StateCommandOption:
		c.currentCommand.Options.Set(c.iterator.Value())
	case StateFlagEqual, StateCommandFlag:
		resp.Error = ErrNotFinished
	}

	return &resp
}

func (c *commandExpression) handleSlash(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	if c.state != StateCommandStart || c.started {
		return resp.WithError(ErrWrongRune)
	}

	c.iterator = ctx.CommandTree().GetIterator()
	c.started = true
	return resp.WithObject(ObjectPath)
}

func (c *commandExpression) handleColon(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	if c.state != StateCommandStart || c.started {
		return resp.WithError(ErrWrongRune)
	}

	c.iterator = ctx.CommandTree().GetIterator()
	c.setRelativeRoot = true
	c.started = true
	return resp.WithObject(ObjectPath)
}

func (c *commandExpression) handleSemicolon(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch c.state {
	case StateCommandPath:
		if !c.iterator.GoToEnd() {
			return resp.WithError(ErrWrongRune)
		}
	case StateCommandCommand:
		if !c.iterator.GoToEnd() {
			return resp.WithError(ErrWrongRune)
		}

		c.closeCommand(resp)
	case StateFlagValue:
		c.flags.Set(c.currentFlag)
	case StateCommandOption:
		c.currentCommand.Options.Set(c.iterator.Value())
	}

	return c.goOut(ctx, resp).WithObject(ObjectOperator)
}

func (c *commandExpression) handleSpace(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	if c.prevRune.Is(' ') {
		return resp.WithObject(ObjectSpace)
	}

	switch c.state {
	case StateCommandPath:
		if !c.iterator.GoToEnd() {
			return resp.WithError(ErrWrongRune)
		}
		nextTree := c.iterator.NextTree()
		if nextTree == nil {
			return resp.WithError(ErrPanic)
		}
		c.iterator = nextTree.GetIterator()
		c.state = StateCommandStart
	case StateCommandCommand:
		if !c.iterator.GoToEnd() {
			return resp.WithError(ErrWrongRune)
		}
		c.state = StateCommandArgument
		return c.closeCommand(resp).WithObject(ObjectSpace)
	case StateCommandFlag:
		return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
	case StateFlagValue:
		c.flags.Set(c.currentFlag)
		c.flagTree.Use(c.currentFlag.Name)
		c.iterator = c.flagTree.GetIterator()
		c.state = StateCommandArgument
	case StateCommandOption:
		if !c.iterator.GoToEnd() {
			return resp.WithError(ErrWrongRune)
		}

		optionName := c.iterator.Value()
		c.currentCommand.Options.Set(optionName)
		c.flagTree.Use(optionName)
		c.iterator = c.flagTree.GetIterator()
		c.state = StateCommandArgument
	case StateFlagEqual:
		return resp.WithError(ErrWrongRune)
	}

	return resp.WithObject(ObjectSpace)
}

func (c *commandExpression) handleLowerAlpha(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch c.state {
	case StateCommandStart:
		if !c.iterator.GoNext(r) {
			return resp.WithError(ErrWrongRune)
		}

		options := c.iterator.NextOptions()

		if options.AggregatedLevel == LevelTypeCommand {
			c.state = StateCommandCommand
			resp.WithObject(ObjectCommand)
		} else {
			c.state = StateCommandPath
			resp.WithObject(ObjectPath)
		}
	case StateCommandPath:
		if !c.iterator.GoNext(r) {
			return resp.WithError(ErrWrongRune)
		}
		resp.WithObject(ObjectPath)
	case StateCommandCommand:
		if !c.iterator.GoNext(r) {
			return resp.WithError(ErrWrongRune)
		}
		resp.WithObject(ObjectCommand)
	case StateCommandArgument:
		if c.prevRune.IsSpace() {
			c.unnamedFlagPosition++
		}

		options := c.iterator.NextOptions()

		// check unnamed flag
		if len(options.Options) == 0 || !c.iterator.GoNext(r) {
			return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
		}

		switch options.AggregatedLevel {
		case LevelTypeFlag:
			c.state = StateCommandFlag
			resp.WithObject(ObjectOptionalFlag)
		case LevelTypeOption:
			c.state = StateCommandOption
			resp.WithObject(ObjectOption)
		default:
			resp.WithObject(ObjectUnknown)
		}

		c.unnamedFlagValue = append(c.unnamedFlagValue, r)
	case StateCommandFlag:
		if !c.iterator.GoNext(r) {
			return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
		}
		c.unnamedFlagValue = append(c.unnamedFlagValue, r)
		resp.WithObject(ObjectOptionalFlag)
	case StateCommandOption:
		if !c.iterator.GoNext(r) {
			return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
		}
		c.unnamedFlagValue = append(c.unnamedFlagValue, r)
		resp.WithObject(ObjectOption)
	case StateFlagEqual:
		c.state = StateFlagValue
		resp.WithObject(ObjectValue)
		return c.addFlag(NewStdExpression(false), resp)
	default:
		return resp.WithError(fmt.Errorf("unknown state %v", c.state))
	}

	return resp
}

func (c *commandExpression) handleNumber(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch c.state {
	case StateCommandArgument:
		if c.prevRune.IsSpace() {
			c.unnamedFlagPosition++
		}
		resp.WithObject(ObjectValue)
		return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
	case StateFlagEqual:
		switch c.currentFlag.ValueType {
		case ValueTypeString:
			c.state = StateFlagValue
			resp.WithObject(ObjectValue)
			return c.addFlag(NewStdExpression(false), resp)
		case ValueTypeNumber:
			c.state = StateFlagValue
			resp.WithObject(ObjectValue)
			return c.addFlag(NewStdExpression(false), resp)
		default:
			return resp.WithError(ErrWrongRune)
		}
	}

	return resp
}

func (c *commandExpression) handleQuote(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch c.state {
	case StateCommandArgument:
		if c.prevRune.IsSpace() {
			c.unnamedFlagPosition++
		}
		c.quoteOpened++
		resp.WithObject(ObjectQuotedSymbol)
		return c.checkUnnamedFlag(ctx, NewStdExpression(false), resp)
	case StateFlagEqual:
		c.state = StateFlagValue
		c.quoteOpened++
		resp.WithObject(ObjectQuotedSymbol)
		return c.addFlag(NewStdExpression(false), resp)
	case StateFlagValue:
		if c.quoteOpened <= 0 {
			return resp.WithError(ErrWrongRune)
		}
		c.quoteOpened--
		resp.WithObject(ObjectQuotedSymbol)
	default:
		return resp.WithError(ErrWrongRune)
	}

	return resp
}

func (c *commandExpression) handleVariable(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch c.state {
	case StateCommandArgument:
		if c.prevRune.IsSpace() {
			c.unnamedFlagPosition++
		}
		resp.WithObject(ObjectVariableSymbol)
		return c.checkUnnamedFlag(ctx, NewVariable(false), resp)
	case StateFlagEqual:
		c.state = StateFlagValue
		resp.WithObject(ObjectVariableSymbol)
		return c.addFlag(NewVariable(false), resp)
	default:
		return resp.WithError(ErrWrongRune)
	}
}

func (c *commandExpression) handleEqual() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	if !c.iterator.GoToEnd() {
		return resp.WithError(ErrWrongRune)
	}

	if c.iterator.Level() != LevelTypeFlag {
		return resp.WithError(ErrWrongRune)
	}

	if c.state != StateCommandFlag {
		return resp.WithError(ErrWrongRune)
	}

	flag, ok := c.iterator.Payload().(*Flag)
	if !ok {
		return resp.WithError(ErrWrongPayload)
	}

	c.currentFlag = flag.Copy()
	c.iterator = c.flagTree.GetIterator()

	c.state = StateFlagEqual
	c.flagUsed = true
	return resp.WithObject(ObjectEqualSymbol)
}

func (c *commandExpression) handleCommandList(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectSquareBrackets)

	switch {
	case r.Is('[') || r.Is('{'):
		isCurly := r.Is('{')

		switch c.state {
		case StateCommandArgument:
			if c.prevRune.IsSpace() {
				c.unnamedFlagPosition++
			}
			c.opened++

			return c.checkUnnamedFlag(ctx, NewCommandList(false, isCurly), resp)
		case StateFlagEqual:
			c.state = StateFlagValue
			c.opened++
			return c.addFlag(NewCommandList(false, isCurly), resp)
		default:
			return resp.WithError(ErrWrongRune)
		}
	case r.Is(']') || r.Is('}'):
		switch c.state {
		case StateFlagValue:
			c.flags.Set(c.currentFlag)
			c.flagTree.Use(c.currentFlag.Name)
			c.iterator = c.flagTree.GetIterator()
			c.state = StateCommandArgument
		case StateCommandOption:
			if !c.iterator.GoToEnd() {
				return resp.WithError(ErrWrongRune)
			}
			c.currentCommand.Options.Set(c.iterator.Value())
		}

		// we don't need to decrease the counter to go next
		if c.opened == 0 {
			return c.goOut(ctx, resp)
		}
	}

	return resp
}

func (c *commandExpression) handleMath(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectRoundBrackets)

	switch {
	case r.Is('('):
		switch c.state {
		case StateCommandArgument:
			if c.prevRune.IsSpace() {
				c.unnamedFlagPosition++
			}
			c.opened++
			return c.checkUnnamedFlag(ctx, NewMathExpression(), resp)
		case StateFlagEqual:
			c.state = StateFlagValue
			c.opened++
			return c.addFlag(NewMathExpression(), resp)
		default:
			return resp.WithError(ErrWrongRune)
		}
	case r.Is(')'):
		switch c.state {
		case StateFlagValue:
			c.flags.Set(c.currentFlag)
			c.flagTree.Use(c.currentFlag.Name)
			c.iterator = c.flagTree.GetIterator()
			c.state = StateCommandArgument
		case StateCommandOption:
			if !c.iterator.GoToEnd() {
				return resp.WithError(ErrWrongRune)
			}
			c.currentCommand.Options.Set(c.iterator.Value())
		}

		// we don't need to decrease the counter to go next
		if c.opened == 0 {
			return c.goOut(ctx, resp)
		}
	}

	return resp
}

func (c *commandExpression) closeCommand(resp *Response) *Response {
	cmd, ok := c.iterator.Payload().(*Command)
	if !ok {
		return resp.WithError(ErrWrongPayload)
	}
	if cmd != nil {
		c.currentCommand = cmd.Copy()
	}
	c.flagTree = c.iterator.NextTree().Copy()
	c.iterator = c.flagTree.GetIterator()
	return resp
}

func (c *commandExpression) goOut(ctx SystemContext, resp *Response) *Response {
	if c.currentCommand != nil {
		c.currentCommand.Out(ctx, c.flags)
	}

	return resp.WithAction(ResponseGoOut)
}

func (c *commandExpression) checkUnnamedFlag(ctx SystemContext, exp Expression, resp *Response) *Response {
	if c.flagUsed {
		return resp.WithError(ErrWrongRune)
	}

	flag := c.currentCommand.UnnamedFlag(c.unnamedFlagPosition)
	if flag == nil {
		return resp.WithError(ErrWrongRune)
	}

	c.currentFlag = flag.Copy()
	c.currentFlag.Set(exp)

	for _, r := range c.unnamedFlagValue {
		exp.Add(ctx, r)
	}

	c.unnamedFlagValue = c.unnamedFlagValue[:0]
	c.state = StateFlagValue

	return resp.WithAction(ResponseRepeat).WithExpression(exp).WithObject(ObjectValue)
}

func (c *commandExpression) addFlag(exp Expression, resp *Response) *Response {
	c.currentFlag.Set(exp)
	return resp.WithAction(ResponseRepeat).WithExpression(exp)
}
