package parser

import "github.com/blkmlk/microshell/internal/models"

type ResponseAction int

const (
	ResponseGoNext ResponseAction = iota
	ResponseGoOut
	ResponseRepeat
)

type ContextType int

const (
	ContextTypeNone ContextType = iota
	ContextTypeCopied
	// TODO: Rename it
	ContextTypeNew
)

type Object int

const (
	ObjectNone Object = iota
	ObjectError
	ObjectPath
	ObjectCommand
	ObjectMandatoryFlag
	ObjectOptionalFlag
	ObjectUnknown
	ObjectValue
	ObjectOption
	ObjectVariableName
	ObjectVariableWrongName
	ObjectQuotedString
	ObjectComment

	// symbols
	ObjectSpace
	ObjectEqualSymbol
	ObjectVariableSymbol
	ObjectQuotedSymbol
	ObjectOperator
	ObjectSquareBrackets
	ObjectRoundBrackets
	ObjectCurlyBrackets
)

func (o Object) IsSingle() bool {
	return o >= ObjectSpace
}

type CloseResponse struct {
	UnclosedBrackets models.Rune
	Error            error
}

type Response struct {
	exp     Expression
	action  ResponseAction
	ctxType ContextType
	object  Object
	err     error
}

func (r *Response) Expression() Expression {
	return r.exp
}

func (r *Response) Err() error {
	return r.err
}

func (r *Response) Action() ResponseAction {
	return r.action
}

func (r *Response) ContextType() ContextType {
	return r.ctxType
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) WithExpression(exp Expression) *Response {
	r.exp = exp
	return r
}

func (r *Response) WithAction(action ResponseAction) *Response {
	r.action = action
	return r
}

func (r *Response) WithObject(object Object) *Response {
	r.object = object
	return r
}

func (r *Response) WithError(err error) *Response {
	r.err = err
	return r
}

func (r *Response) WithContextType(ctxType ContextType) *Response {
	r.ctxType = ctxType
	return r
}
