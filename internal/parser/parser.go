package parser

import (
	"context"
	"errors"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"

	"github.com/blkmlk/microshell/internal/models"
)

type ParsedObject struct {
	Object Object
	Length int
}

type ParseRuneResponse struct {
	Object Object
}

type ParseStringResponse struct {
	Objects []*ParsedObject
	Error   error
}

type ExecResponse struct {
	UnclosedBrackets models.Rune
	Error            error
	Value            Value
}

type parser struct {
	logger          logger.Logger
	rootCtx         SystemContext
	currentCtx      SystemContext
	currentCancel   context.CancelFunc
	expressionStack *ExpressionStack
}

func newParser(ctn di.Container) Parser {
	parser := new(parser)
	parser.logger = ctn.Get(logger.DefinitionName).(logger.Logger)
	parser.rootCtx = ctn.Get(DefinitionNameRootScope).(SystemContext)
	parser.Flush()

	return parser
}

func (p *parser) Flush() {
	ctx, cancel := context.WithCancel(p.rootCtx.Ctx())
	p.currentCtx = p.rootCtx.New().WithContext(ctx)
	p.currentCancel = cancel
	p.expressionStack = newExpressionStack()
	p.expressionStack.Push(p.currentCtx, NewCommandList(true, false))
}

func (p *parser) IsFlushed() bool {
	panic("implement me")
}

func (p *parser) Add(r models.Rune) (*ParseRuneResponse, error) {
	ctx, exp := p.expressionStack.Pop()

	if exp == nil {
		return nil, errors.New("exp nil")
	}

	resp := exp.Add(ctx, r)

	if resp.Err() != nil {
		return nil, resp.Err()
	}

	if resp.Expression() != nil && resp.Expression() != exp && resp.Action() != ResponseGoOut {
		p.expressionStack.Push(ctx, exp)
		exp = resp.Expression()
	}

	switch resp.ContextType() {
	case ContextTypeNew:
		ctx = ctx.New()
	case ContextTypeCopied:
		ctx = ctx.Copy()
	}

	switch resp.Action() {
	case ResponseGoNext:
		p.expressionStack.Push(ctx, exp)
	case ResponseRepeat:
		p.expressionStack.Push(ctx, exp)
		return p.Add(r)
	case ResponseGoOut:
		if p.expressionStack.Size() == 0 {
			return nil, errors.New("can't go out")
		}

		closeResp := exp.Close(ctx)

		if resp != nil && closeResp.Error != nil {
			return nil, closeResp.Error
		}

		return p.Add(r)
	}

	return &ParseRuneResponse{
		Object: resp.object,
	}, nil
}

func (p *parser) ParseString(s string) *ParseStringResponse {
	p.Flush()
	var (
		response ParseStringResponse
		newObj   Object
		err      error
	)

	var obj = new(ParsedObject)
	obj.Object = ObjectSpace

	for _, c := range s {
		r := models.Rune(c)

		if err == nil {
			resp, inErr := p.Add(r)

			if inErr != nil {
				err = inErr

				if r.IsSpace() {
					obj.Object = ObjectError
					newObj = ObjectSpace
				} else {
					newObj = ObjectError
				}
			} else {
				newObj = resp.Object
			}
		} else {
			switch {
			case r.IsSpace():
				newObj = ObjectSpace
			case obj.Object == ObjectSpace:
				newObj = ObjectNone
			default:
				newObj = obj.Object
			}
		}

		if newObj != obj.Object {
			if obj.Object.IsSingle() || newObj.IsSingle() {
				if newObj == ObjectSpace &&
					(obj.Object == ObjectOptionalFlag || obj.Object == ObjectMandatoryFlag || obj.Object == ObjectUnknown) {
					obj.Object = ObjectValue
				}

				response.Objects = append(response.Objects, obj)
				obj = new(ParsedObject)
			}

			obj.Object = newObj
		}

		obj.Length++
	}

	if obj.Object != ObjectNone {
		response.Objects = append(response.Objects, obj)
	}

	response.Error = err

	return &response
}

func (p *parser) Exec() (*ExecResponse, error) {
	if p.expressionStack.Size() == 0 {
		return nil, errors.New("no expression")
	}

	var (
		ctx  SystemContext
		exp  Expression
		resp ExecResponse
	)

	for p.expressionStack.Size() != 0 {
		ctx, exp = p.expressionStack.Pop()

		closeResp := exp.Close(ctx)

		if closeResp.Error != nil {
			resp.Error = closeResp.Error
			resp.UnclosedBrackets = closeResp.UnclosedBrackets
			return &resp, nil
		}
	}

	resp.Value = exp.Value(ctx)

	return &resp, nil
}

func (p *parser) Continue() *CompleteResponse {
	ctx, exp := p.expressionStack.Pop()
	defer p.expressionStack.Push(ctx, exp)

	return exp.Complete(ctx)
}
