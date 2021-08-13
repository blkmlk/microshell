package parser

import (
	"testing"

	"github.com/blkmlk/microshell/internal/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/blkmlk/microshell/internal/logger"
	"github.com/blkmlk/microshell/internal/models"

	"github.com/sarulabs/di/v2"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
	ctn    di.Container
	parser *parser
}

func TestParser(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (t *ParserTestSuite) SetupTest() {
	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(
		Definition,
		DefinitionContext,
		DefinitionScope,
		DefinitionCommandTree,
		logger.Definition,
	)
	t.Require().NoError(err)
	t.ctn = builder.Build()

	t.parser = t.ctn.Get(DefinitionName).(*parser)
}

func (t *ParserTestSuite) TestParser() {
	ctx, err := newRootContext(t.ctn)
	t.Require().NoError(err)

	exp1 := new(MockExpression)
	exp2 := new(MockExpression)
	exp3 := new(MockExpression)

	t.parser.expressionStack.Push(ctx, exp1)
	t.parser.expressionStack.Push(ctx, exp2)
	t.parser.expressionStack.Push(ctx, exp3)
	t.Require().Equal(4, t.parser.expressionStack.Size())

	// go next
	invoked := 0
	exp3.On("Add", mocks.AnyArgument, models.Rune('a')).
		Run(func(args mock.Arguments) {
			invoked = 1
		}).
		Return(NewResponse().WithAction(ResponseGoNext)).Once()
	_, err = t.parser.Add(models.Rune('a'))
	t.Require().NoError(err)
	t.Require().Equal(1, invoked)
	t.Require().Equal(4, t.parser.expressionStack.Size())

	// repeat
	invoked = 0
	exp3.On("Add", mocks.AnyArgument, models.Rune('a')).
		Run(func(args mock.Arguments) {
			invoked++
		}).
		Return(NewResponse().WithAction(ResponseRepeat)).Once()
	exp3.On("Add", mocks.AnyArgument, models.Rune('a')).
		Run(func(args mock.Arguments) {
			invoked++
		}).
		Return(NewResponse().WithAction(ResponseGoNext)).Once()
	_, err = t.parser.Add(models.Rune('a'))
	t.Require().NoError(err)
	t.Require().Equal(2, invoked)
	t.Require().Equal(4, t.parser.expressionStack.Size())

	// go out
	invoked = 0
	exp3.On("Add", mocks.AnyArgument, models.Rune('a')).
		Run(func(args mock.Arguments) {
			invoked++
		}).
		Return(NewResponse().WithAction(ResponseGoOut)).Once()
	exp3.On("Close", mocks.AnyArgument).Return(&CloseResponse{}, nil).Once()
	exp2.On("Add", mocks.AnyArgument, models.Rune('a')).
		Run(func(args mock.Arguments) {
			invoked++
		}).
		Return(NewResponse().WithAction(ResponseGoNext)).Once()
	_, err = t.parser.Add(models.Rune('a'))
	t.Require().NoError(err)
	t.Require().Equal(2, invoked)
	t.Require().Equal(3, t.parser.expressionStack.Size())
}
