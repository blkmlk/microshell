package parser

import (
	"strconv"
	"strings"

	"github.com/blkmlk/microshell/internal/models"
)

var (
	boolValueTrue  = []models.Rune{'t', 'r', 'u', 'e'}
	boolValueFalse = []models.Rune{'f', 'a', 'l', 's', 'e'}
)

type stdExpression struct {
	strictMode bool
	quotes     int

	boolValueIdx int
	boolValue    []models.Rune

	value strings.Builder
}

func NewStdExpression(strictMode bool) Expression {
	return &stdExpression{strictMode: strictMode}
}

func (s *stdExpression) Type() ExpressionType {
	return ExpressionTypeStd
}

func NewStringValue(v string) Value {
	exp := new(stdExpression)
	exp.value.WriteString(v)

	return exp
}

func NewNumberValue(v int) Value {
	return NewStringValue(strconv.Itoa(v))
}

func NewBoolValue(value bool) Value {
	if value {
		return NewStringValue("true")
	}
	return NewStringValue("false")
}

func (s *stdExpression) Complete(ctx SystemContext) *CompleteResponse {
	return &CompleteResponse{Merged: " "}
}

func (s *stdExpression) Add(ctx SystemContext, r models.Rune) *Response {
	var resp = NewResponse().WithAction(ResponseGoNext).WithObject(ObjectValue)

	switch {
	case r.IsAlpha():
		if s.quotes >= 2 {
			return resp.WithObject(ObjectValue)
		}

		if s.quotes == 1 {
			s.value.WriteRune(rune(r))
			return resp.WithObject(ObjectQuotedString)
		}

		if !s.strictMode {
			s.value.WriteRune(rune(r))
			return resp.WithObject(ObjectValue)
		}

		if s.boolValueIdx == 0 {
			if r.Is('t') {
				s.boolValue = boolValueTrue
			} else if r.Is('f') {
				s.boolValue = boolValueFalse
			} else {
				return resp.WithError(ErrWrongRune)
			}

			s.boolValueIdx++
			return resp.WithObject(ObjectValue)
		}

		if s.boolValueIdx >= len(s.boolValue) || s.boolValue[s.boolValueIdx] != r {
			return resp.WithError(ErrWrongRune)
		}

		s.boolValueIdx++

		if s.boolValueIdx == len(s.boolValue) {
			for _, c := range s.boolValue {
				s.value.WriteRune(rune(c))
			}
		}

		return resp.WithObject(ObjectValue)
	case r.IsNumber():
		resp.WithObject(ObjectValue)
		s.value.WriteRune(rune(r))
	case r.IsSpace():
		resp.WithObject(ObjectSpace)

		if s.quotes == 1 {
			s.value.WriteRune(rune(r))
			break
		}

		if s.value.Len() == 0 {
			return resp.WithError(ErrWrongRune)
		}

		if s.strictMode && s.boolValueIdx != len(s.boolValue) {
			return resp.WithError(ErrWrongRune)
		}

		return resp.WithAction(ResponseGoOut)
	case r.Is('"'):
		resp.WithObject(ObjectQuotedSymbol)
		s.quotes++

		if s.quotes == 2 {
			return resp.WithAction(ResponseGoOut)
		}

		if s.value.Len() != 0 {
			return resp.WithError(ErrWrongRune)
		}

		if s.quotes > 2 {
			return resp.WithError(ErrWrongRune)
		}
	default:
		return resp.WithAction(ResponseGoOut)
	}

	return resp
}

func (s *stdExpression) Value(ctx SystemContext) Value {
	return s
}

func (s *stdExpression) Close(ctx SystemContext) *CloseResponse {
	var resp CloseResponse

	if s.quotes == 1 {
		resp.UnclosedBrackets = '"'
		resp.Error = ErrNotFinished
		return &resp
	}

	if s.boolValueIdx != len(s.boolValue) {
		resp.Error = ErrWrongRune
		return &resp
	}

	return &resp
}

func (s *stdExpression) Bool() bool {
	return s.value.String() == "true"
}

func (s *stdExpression) Number() int {
	if s.value.Len() == 0 {
		return 0
	}
	v, _ := strconv.Atoi(s.value.String())
	return v
}

func (s *stdExpression) String() string {
	return s.value.String()
}

func (s *stdExpression) IsBool() bool {
	v := s.value.String()
	return v == "true" || v == "false"
}

func (s *stdExpression) IsString() bool {
	return true
}

func (s stdExpression) IsNumber() bool {
	if s.value.Len() == 0 {
		return true
	}

	_, err := strconv.Atoi(s.value.String())
	return err == nil
}

func (s *stdExpression) Equal(v Value) bool {
	return s.String() == v.String()
}

func (s *stdExpression) Less(v Value) bool {
	switch {
	case s.IsBool():
		return false
	case s.IsNumber():
		return s.Number() < v.Number()
	case s.IsString():
		return strings.Compare(s.String(), v.String()) == -1
	}

	return false
}

func (s stdExpression) Greater(v Value) bool {
	switch {
	case s.IsBool():
		return false
	case s.IsNumber():
		return s.Number() > v.Number()
	case s.IsString():
		return strings.Compare(s.String(), v.String()) == 1
	}

	return false
}
