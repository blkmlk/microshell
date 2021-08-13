package parser

import (
	"github.com/blkmlk/microshell/internal/models"
)

type commandList struct {
	expressions     []Expression
	innerExpression Expression
	value           Value
	rootMode        bool
	listRune        models.Rune
	usedRunes       map[models.Rune]int
	closed          bool
}

func NewCommandList(rootMode, isCurly bool) Expression {
	cl := &commandList{
		rootMode:  rootMode,
		usedRunes: make(map[models.Rune]int),
	}

	if isCurly {
		cl.listRune = '{'
	} else {
		cl.listRune = '['
	}

	if rootMode {
		cl.usedRunes[cl.listRune]++
	}

	return cl
}

func (c *commandList) Type() ExpressionType {
	return ExpressionTypeCmdList
}

func (c *commandList) Complete(ctx SystemContext) *CompleteResponse {
	return NewCommandExpression(ctx).Complete(ctx)
}

func (c *commandList) Add(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	switch {
	case r.Is(';'):
		if c.innerExpression == nil {
			return resp.WithAction(ResponseGoNext).WithObject(ObjectOperator).WithObject(ObjectOperator)
		}

		c.innerExpression.Close(ctx)

		c.expressions = append(c.expressions, c.innerExpression)
		c.innerExpression = nil
		return resp.WithAction(ResponseGoNext).WithObject(ObjectOperator)
	case r.Is('/') || r.Is(':') || r.IsLowerAlpha():
		if c.innerExpression == nil {
			c.innerExpression = NewCommandExpression(ctx)
		}

		resp.WithAction(ResponseRepeat).WithExpression(c.innerExpression)
	case r.Is(' '):
		return resp.WithAction(ResponseGoNext).WithObject(ObjectSpace)
	case r.Is('$'):
		c.innerExpression = NewVariable(true)
		resp.WithAction(ResponseRepeat).WithExpression(c.innerExpression)
	case r.Is('['), r.Is('{'), r.Is('('):
		return c.openRune(r)
	case r.Is(']'), r.Is('}'), r.Is(')'):
		return c.closeRune(r)
	default:
		return resp.WithError(ErrWrongRune)
	}

	return resp
}

func (c *commandList) Close(ctx SystemContext) *CloseResponse {
	var resp = new(CloseResponse)

	for r, n := range c.usedRunes {
		if n <= 0 || (c.rootMode && n == 1) {
			continue
		}

		resp.UnclosedBrackets = r
		resp.Error = ErrNotFinished
		return resp
	}

	if !c.closed {
		if c.innerExpression != nil {
			c.expressions = append(c.expressions, c.innerExpression)
			c.innerExpression = nil
		}
	}

	c.closed = true
	return resp
}

func (c *commandList) Value(ctx SystemContext) Value {
	var value Value

	if c.listRune.Is('{') {
		ctx = ctx.New()
	}

	for _, expr := range c.expressions {
		value = expr.Value(ctx)

		if _, ok := value.(error); ok {
			// TODO: handle err
			return NullValue
		}
	}

	return value
}

func (c *commandList) openRune(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	defer func() {
		c.usedRunes[r]++
	}()

	if c.usedRunes[r] > 0 || c.listRune != r {
		var ctxType ContextType
		switch r {
		case '[':
			c.innerExpression = NewCommandList(false, false)
			ctxType = ContextTypeCopied
		case '{':
			c.innerExpression = NewCommandList(false, true)
			ctxType = ContextTypeNew
		case '(':
			c.innerExpression = NewMathExpression()
			ctxType = ContextTypeNone
		default:
			return resp.WithError(ErrWrongRune)
		}

		return resp.WithAction(ResponseRepeat).WithExpression(c.innerExpression).WithContextType(ctxType)
	}

	var obj = ObjectSquareBrackets
	switch r {
	case '{':
		obj = ObjectCurlyBrackets
	case '(':
		obj = ObjectRoundBrackets
	}

	return resp.WithAction(ResponseGoNext).WithObject(obj)
}

func (c *commandList) closeRune(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext)

	var openRune models.Rune
	switch r {
	case ']':
		openRune = '['
	case '}':
		openRune = '{'
	case ')':
		openRune = '('
	}

	if c.listRune != openRune && c.usedRunes[openRune] != 1 {
		return resp.WithError(ErrWrongRune)
	}

	if c.listRune == openRune && c.usedRunes[openRune] < 1 {
		return resp.WithError(ErrWrongRune)
	}

	if c.rootMode && c.listRune == openRune && c.usedRunes[openRune] == 1 {
		return resp.WithError(ErrWrongRune)
	}

	if c.innerExpression != nil {
		c.expressions = append(c.expressions, c.innerExpression)
		c.innerExpression = nil
	}

	c.usedRunes[openRune]--

	if c.listRune != openRune || c.usedRunes[openRune] == 1 {
		var obj = ObjectSquareBrackets
		switch r {
		case '}':
			obj = ObjectCurlyBrackets
		case ')':
			obj = ObjectRoundBrackets
		}

		return resp.WithAction(ResponseGoNext).WithObject(obj)
	}

	c.closed = true
	return resp.WithAction(ResponseGoOut)
}
