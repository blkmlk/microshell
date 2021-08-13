package parser

import (
	"github.com/blkmlk/microshell/internal/models"
)

type VariableState int

const (
	VariableStateStart VariableState = iota
	VariableStateName
	VariableStateArgument
	VariableStateFlagName
	VariableStateValue
)

type variableArgument struct {
	Name       string
	Expression Expression
}

type variableExpression struct {
	name            string
	state           VariableState
	iterator        *variableIterator
	arguments       []*variableArgument
	currentArgument *variableArgument
	clOpened        int
	mathOpened      int
	functionMode    bool
}

func NewVariable(functionMode bool) Expression {
	return &variableExpression{
		state:        VariableStateStart,
		functionMode: functionMode,
	}
}

func (v *variableExpression) Type() ExpressionType {
	return ExpressionTypeVar
}

func (v *variableExpression) Complete(ctx SystemContext) *CompleteResponse {
	if v.state != VariableStateName || v.iterator == nil {
		return nil
	}

	opts := v.iterator.NextOptions()

	var options = make([]*CompleteOption, 0, len(opts.Options))
	for _, o := range opts.Options {
		options = append(options, &CompleteOption{
			Level:  o.Level,
			Option: o.Name,
		})
	}

	return &CompleteResponse{
		Options: options,
		Merged:  opts.Merged,
	}
}

func (v *variableExpression) Add(ctx SystemContext, r models.Rune) *Response {
	var resp *Response

	switch {
	case r.Is('$'):
		resp = v.handleVarSign(ctx)
	case r.IsNumber() || r.IsAlpha():
		resp = v.handleAlpha(ctx, r)
	case r.Is('='):
		resp = v.handleEqual()
	case r.Is('[') || r.Is(']'):
		resp = v.handleCommandList(r)
	case r.Is('(') || r.Is(')'):
		resp = v.handleMath(r)
	case r.IsSpace():
		resp = v.handleSpace()
	default:
		return NewResponse().WithAction(ResponseGoOut)
	}

	return resp
}

func (v *variableExpression) handleVarSign(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectVariableSymbol)

	if v.state == VariableStateValue {
		v.currentArgument.Expression = NewVariable(false)
		return resp.WithAction(ResponseRepeat).WithExpression(v.currentArgument.Expression)
	}

	if v.state != VariableStateStart {
		return resp.WithError(ErrWrongRune)
	}

	v.iterator = ctx.VariableTree().GetIterator()
	v.state = VariableStateName

	return resp
}

func (v *variableExpression) handleAlpha(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectVariableWrongName)

	switch v.state {
	case VariableStateName:
		v.name += r.String()

		if ctx.VariableExists(v.name) {
			resp.WithObject(ObjectVariableName)
		} else {
			resp.WithObject(ObjectVariableWrongName)
		}
		v.iterator.Next(r)
	case VariableStateArgument, VariableStateFlagName:
		v.currentArgument.Name += r.String()

		// TODO: change it
		v.state = VariableStateFlagName
		resp.WithObject(ObjectVariableName)
	case VariableStateValue:
		v.currentArgument.Expression = NewStdExpression(false)
		return resp.WithAction(ResponseRepeat).WithExpression(v.currentArgument.Expression)
	default:
		return resp.WithError(ErrWrongRune)
	}

	return resp
}

func (v *variableExpression) handleSpace() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectSpace)

	if !v.functionMode {
		return resp.WithAction(ResponseGoOut)
	}

	switch v.state {
	case VariableStateName, VariableStateValue:
		if v.currentArgument != nil {
			v.arguments = append(v.arguments, v.currentArgument)
		}

		v.currentArgument = new(variableArgument)
		v.state = VariableStateArgument
	case VariableStateArgument:
	default:
		return resp.WithError(ErrWrongRune)
	}

	return resp
}

func (v *variableExpression) handleEqual() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectEqualSymbol)

	if v.state != VariableStateFlagName {
		return resp.WithError(ErrWrongRune)
	}

	v.state = VariableStateValue

	return resp
}

func (v *variableExpression) handleCommandList(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectSquareBrackets)

	switch {
	case r.Is('['):
		if v.state != VariableStateValue {
			return resp.WithError(ErrWrongRune)
		}

		v.clOpened++
		v.currentArgument.Expression = NewCommandList(false, false)
		return resp.WithAction(ResponseRepeat).WithExpression(v.currentArgument.Expression)
	case r.Is(']'):
		if v.clOpened == 0 {
			return resp.WithAction(ResponseGoOut)
		}
		v.clOpened--
	}

	return resp
}

func (v *variableExpression) handleMath(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectRoundBrackets)

	switch {
	case r.Is('('):
		if v.state != VariableStateValue {
			return resp.WithError(ErrWrongRune)
		}

		v.mathOpened++
		v.currentArgument.Expression = NewMathExpression()
		return resp.WithAction(ResponseRepeat).WithExpression(v.currentArgument.Expression)
	case r.Is(')'):
		if v.mathOpened == 0 {
			return resp.WithAction(ResponseGoOut)
		}
		v.mathOpened--
	}

	return resp
}

func (v *variableExpression) Value(ctx SystemContext) Value {
	if v.iterator == nil {
		return NullValue
	}

	payload := ctx.GetVariablePayload(v.name)

	if payload == nil {
		return NullValue
	}

	if value, ok := payload.(Value); ok {
		return value
	}

	if _, ok := payload.(Valuer); !ok {
		return NullValue
	}

	valuer := payload.(Valuer)

	if v.functionMode {
		innerCtx := ctx.New()

		for _, arg := range v.arguments {
			innerCtx.SetLocalVariable(arg.Name, arg.Expression)
		}

		return valuer.Value(innerCtx)
	}

	return valuer.Value(ctx)
}

func (v *variableExpression) Close(ctx SystemContext) *CloseResponse {
	var resp CloseResponse

	if v.name == "" {
		resp.Error = ErrNotFinished
		return &resp
	}

	if v.state == VariableStateValue {
		if v.currentArgument != nil {
			v.arguments = append(v.arguments, v.currentArgument)
		}
	}

	return &resp
}

func (v *variableExpression) Debug() string {
	return ""
}
