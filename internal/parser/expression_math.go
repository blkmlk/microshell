package parser

import (
	"fmt"

	"github.com/blkmlk/microshell/internal/models"
)

const (
	StateMathNone MathState = iota
	StateMathOperatorAfterNone
	StateMathOperatorNotAfterExpression
	StateMathOperatorNotAfterFinished
	StateMathOperatorFinished
	StateMathOperatorNotFinished
	StateMathExpression
)

type mathExpression struct {
	state             MathState
	tree              *mathTree
	lastExpression    Expression
	openedQuoted      bool
	openedBrackets    int
	opened            int
	lastOperator      Operator
	prevRune          models.Rune
	expressionBalance int
	completed         bool
}

type MathState int

func NewMathExpression() Expression {
	return &mathExpression{
		tree: NewMathTree(),
	}
}

func (m *mathExpression) Type() ExpressionType {
	return ExpressionTypeMath
}

func (m *mathExpression) Value(ctx SystemContext) Value {
	v, err := m.tree.Value(ctx)

	if err != nil {
		// TODO: handle it
		//log.Fatal(err)
		fmt.Println(err.Error())
	}

	if v == nil {
		v = NewNumberValue(0)
	}

	return v
}

func (m *mathExpression) Close(ctx SystemContext) *CloseResponse {
	var resp CloseResponse

	if m.state != StateMathExpression {
		resp.Error = ErrNotFinished
		return &resp
	}

	m.tree.Add(m.lastExpression)

	if m.opened > 0 {
		resp.UnclosedBrackets = '('
		resp.Error = ErrNotFinished
	}

	return &resp
}

func (m *mathExpression) Debug() string {
	return ""
}

func (m *mathExpression) Complete(ctx SystemContext) *CompleteResponse {
	return &CompleteResponse{}
}

func (m *mathExpression) Add(ctx SystemContext, r models.Rune) *Response {
	var resp *Response

	m.completed = false

	switch {
	case r.Is(' '):
		resp = m.handleSpace()
	case r.IsNumber():
		resp = m.handleAlpha(ctx)
	case r.IsAlpha():
		resp = m.handleAlpha(ctx)
	case r.Is('"'):
		resp = m.handleQuoteString(ctx)
	case r.Is('!'):
		resp = m.handleUnaryOperator()
	case r.Is('.') || r.Is('+') || r.Is('-') || r.Is('/') || r.Is('*'):
		resp = m.handleOperator(r)
	case r.Is('>') || r.Is('<') || r.Is('='):
		resp = m.handleCompareOperator(r)
	case r.Is('('):
		resp = m.handleOpenBracket()
	case r.Is(')'):
		resp = m.handleCloseBracket()
	case r.Is('[') || r.Is(']'):
		resp = m.handleCommandList(r)
	case r.Is('$'):
		resp = m.handleVariable()
	default:
		return NewResponse().WithError(ErrWrongRune)
	}

	m.prevRune = r

	return resp
}

func (m *mathExpression) handleSpace() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectSpace)

	if m.state == StateMathOperatorNotFinished {
		m.state = StateMathOperatorFinished
	}

	m.completed = true

	return resp
}

func (m *mathExpression) handleAlpha(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectValue)

	if m.state == StateMathExpression || m.state == StateMathOperatorNotAfterExpression {
		return resp.WithError(ErrWrongRune)
	}

	if m.state != StateMathNone {
		m.tree.Add(m.lastOperator)
	}

	m.state = StateMathExpression
	m.lastExpression = NewStdExpression(true)

	return resp.WithAction(ResponseRepeat).WithExpression(m.lastExpression)
}

func (m *mathExpression) handleQuoteString(ctx SystemContext) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectQuotedSymbol)

	if !m.openedQuoted {
		if m.state == StateMathExpression || m.state == StateMathOperatorNotAfterExpression {
			return resp.WithError(ErrWrongRune)
		}

		if m.state != StateMathNone {
			m.tree.Add(m.lastOperator)
		}

		m.state = StateMathExpression
		m.openedQuoted = true

		m.lastExpression = NewStdExpression(false)
		resp.WithAction(ResponseRepeat).WithExpression(m.lastExpression)
	} else {
		m.lastExpression.Close(ctx)
		m.openedQuoted = false
	}

	return resp
}

func (m *mathExpression) handleUnaryOperator() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectOperator)

	switch m.state {
	case StateMathNone:
		m.state = StateMathOperatorAfterNone
	case StateMathExpression:
		m.state = StateMathOperatorNotAfterExpression
		m.tree.Add(m.lastExpression)
	case StateMathOperatorFinished:
		m.state = StateMathOperatorNotAfterFinished
		m.tree.Add(m.lastOperator)
	default:
		m.tree.Add(m.lastOperator)
	}

	m.lastOperator = OperatorNot

	return resp
}

func (m *mathExpression) handleOperator(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectOperator)

	switch m.state {
	case StateMathNone:
		if r.Is('+') || r.Is('-') {
			m.state = StateMathOperatorAfterNone
		} else {
			return resp.WithError(ErrWrongRune)
		}
	case StateMathOperatorFinished:
		if r.Is('+') || r.Is('-') {
			m.tree.Add(m.lastOperator)
		} else {
			return resp.WithError(ErrWrongRune)
		}
	case StateMathExpression:
		m.state = StateMathOperatorFinished
		m.tree.Add(m.lastExpression)
	default:
		return resp.WithError(ErrWrongRune)
	}

	switch r {
	case '+':
		m.lastOperator = OperatorPlus
	case '-':
		m.lastOperator = OperatorMinus
	case '.':
		m.lastOperator = OperatorConcatenate
	case '/':
		m.lastOperator = OperatorDivide
	case '*':
		m.lastOperator = OperatorMultiply
	}

	return resp
}

func (m *mathExpression) handleCompareOperator(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectOperator)

	switch m.state {
	case StateMathOperatorNotAfterExpression:
		if !r.Is('=') || m.prevRune.IsSpace() {
			return resp.WithError(ErrWrongRune)
		}
		m.lastOperator = OperatorNotEqual
		m.state = StateMathOperatorFinished
	case StateMathOperatorNotFinished:
		if m.prevRune.IsSpace() {
			return resp.WithError(ErrWrongRune)
		}
		switch m.lastOperator {
		case OperatorLess:
			m.lastOperator = OperatorLessOrEqual
		case OperatorGreater:
			m.lastOperator = OperatorGreaterOrEqual
		default:
			return resp.WithError(ErrWrongRune)
		}
		m.state = StateMathOperatorFinished
	case StateMathExpression:
		switch r {
		case '>':
			//	TODO: AddGlobal a bit shifting
			m.lastOperator = OperatorGreater
			m.state = StateMathOperatorNotFinished
		case '<':
			//	TODO: AddGlobal a bit shifting
			m.lastOperator = OperatorLess
			m.state = StateMathOperatorNotFinished
		case '=':
			m.lastOperator = OperatorEqual
			m.state = StateMathOperatorFinished
		default:
			return resp.WithError(ErrWrongRune)
		}

		m.tree.Add(m.lastExpression)
	default:
		return resp.WithError(ErrWrongRune)
	}

	return resp
}

func (m *mathExpression) handleOpenBracket() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectRoundBrackets)

	if m.state == StateMathExpression || m.state == StateMathOperatorNotAfterExpression {
		return resp.WithError(ErrWrongRune)
	}

	m.opened++

	if m.opened == 1 {
		return resp
	}

	if m.state != StateMathNone {
		m.tree.Add(m.lastOperator)
	}

	m.state = StateMathExpression

	m.openedQuoted = true
	m.lastExpression = NewMathExpression()
	resp.WithAction(ResponseRepeat).WithExpression(m.lastExpression)

	return resp
}

func (m *mathExpression) handleCloseBracket() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectRoundBrackets)

	if m.state != StateMathNone && m.state != StateMathExpression {
		return resp.WithError(ErrWrongRune)
	}

	m.opened--

	// TODO: Add comment
	if m.openedQuoted {
		m.openedQuoted = false
		return resp
	}

	resp.WithAction(ResponseGoOut)

	return resp
}

func (m *mathExpression) handleCommandList(r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectSquareBrackets)

	switch {
	case r.Is('['):
		if m.state == StateMathExpression || m.state == StateMathOperatorNotAfterExpression {
			return resp.WithError(ErrWrongRune)
		}

		if m.state != StateMathNone {
			m.tree.Add(m.lastOperator)
		}

		m.openedBrackets++
		m.state = StateMathExpression
		m.lastExpression = NewCommandList(false, false)
		resp.WithAction(ResponseRepeat).WithExpression(m.lastExpression)
	case r.Is(']'):
		if m.openedBrackets <= 0 {
			return resp.WithError(ErrWrongRune)
		}

		m.openedBrackets--
	}

	return resp
}

func (m *mathExpression) handleVariable() *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectVariableSymbol)

	if m.state == StateMathExpression || m.state == StateMathOperatorNotAfterExpression {
		return resp.WithError(ErrWrongRune)
	}

	if m.state != StateMathNone {
		m.tree.Add(m.lastOperator)
	}

	m.state = StateMathExpression
	m.lastExpression = NewVariable(false)
	resp.WithAction(ResponseRepeat).WithExpression(m.lastExpression)

	return resp
}
