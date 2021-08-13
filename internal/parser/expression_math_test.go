package parser

import (
	"testing"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"

	"github.com/stretchr/testify/suite"
)

type MathExpressionTestSuite struct {
	suite.Suite
	scope  SystemContext
	parser Parser
}

func TestMath(t *testing.T) {
	suite.Run(t, new(MathExpressionTestSuite))
}

func (t *MathExpressionTestSuite) SetupTest() {
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

	ctn := builder.Build()

	t.scope = ctn.Get(DefinitionNameRootScope).(SystemContext)
	t.parser = ctn.Get(DefinitionName).(Parser)
}

func (t *MathExpressionTestSuite) TestMathExpression() {
	var v Value
	var err error

	ctx := t.scope
	ctx.SetCommandRoot(NewCommandTree())
	ctx.SetCommandTree(NewCommandTree())

	v, err = t.makeExpression(`(9 *5)`)
	t.Require().NoError(err)
	t.Require().Equal(45, v.Number())

	v, err = t.makeExpression(`(3 . 5 + 5)`)
	t.Require().NoError(err)
	t.Require().Equal(310, v.Number())

	v, err = t.makeExpression(`(9 * (3+3))`)
	t.Require().NoError(err)
	t.Require().Equal(54, v.Number())

	v, err = t.makeExpression(`(5-(8 + 9))`)
	t.Require().NoError(err)
	t.Require().Equal(-12, v.Number())

	v, err = t.makeExpression(`(5 = 5)`)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	v, err = t.makeExpression(`(5 != 1)`)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	v, err = t.makeExpression(`(5 < 5)`)
	t.Require().NoError(err)
	t.Require().False(v.Bool())

	v, err = t.makeExpression(`(5 <= 5)`)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	v, err = t.makeExpression(`(9 > 8)`)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	v, err = t.makeExpression(`(! (5 > 3))`)
	t.Require().NoError(err)
	t.Require().False(v.Bool())

	v, err = t.makeExpression(`(!!(5 > 3))`)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	v, err = t.makeExpression(`((5 + 3) * (5 - 3))`)
	t.Require().NoError(err)
	t.Require().Equal(16, v.Number())

	v, err = t.makeExpression(`(+5)`)
	t.Require().NoError(err)
	t.Require().Equal(5, v.Number())

	v, err = t.makeExpression(`(-5)`)
	t.Require().NoError(err)
	t.Require().Equal(-5, v.Number())

	v, err = t.makeExpression(`(- (5 + 3))`)
	t.Require().NoError(err)
	t.Require().Equal(-8, v.Number())

	_, err = t.makeExpression(`(*5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(1+=5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(>= 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(<= 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(< 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(> 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(= 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(!= 5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(1 != /5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(1 ! /5)`)
	t.Require().Error(err)

	_, err = t.makeExpression(`(1 > +5)`)
	t.Require().NoError(err)

	_, err = t.makeExpression(`(1 + !5)`)
	t.Require().NoError(err)

	_, err = t.makeExpression(`(1 / !5)`)
	t.Require().NoError(err)

	v, err = t.makeExpression(`(5 / 2)`)
	t.Require().NoError(err)
	t.Require().Equal(2, v.Number())

	v, err = t.makeExpression(`("5" . 2)`)
	t.Require().NoError(err)
	t.Require().Equal(52, v.Number())

	v, err = t.makeExpression(`("n" . 2)`)
	t.Require().NoError(err)
	t.Require().Equal("n2", v.String())

	// with a variable
	ctx.SetGlobalVariable("var", NewNumberValue(90))
	v, err = t.makeExpression(`($var + 2)`)
	t.Require().NoError(err)
	t.Require().Equal(92, v.Number())

	ctx.SetGlobalVariable("var2", NewNumberValue(50))
	v, err = t.makeExpression(`($var+$var2)`)
	t.Require().NoError(err)
	t.Require().Equal(140, v.Number())

	// with a command
	v, err = t.makeExpression(`([] + 2)`)
	t.Require().NoError(err)
	t.Require().Equal(2, v.Number())

	v, err = t.makeExpression(`( [ ]-2)`)
	t.Require().NoError(err)
	t.Require().Equal(-2, v.Number())

	v, err = t.makeExpression(`(1 = 1)`)
	t.Require().NoError(err)
	t.Require().Equal(true, v.Bool())

	v, err = t.makeExpression(`(true = true)`)
	t.Require().NoError(err)
	t.Require().Equal(true, v.Bool())

	v, err = t.makeExpression(`(1 = 1 = true)`)
	t.Require().NoError(err)
	t.Require().Equal(true, v.Bool())

	v, err = t.makeExpression(`(!!true)`)
	t.Require().NoError(err)
	t.Require().Equal(true, v.Bool())
}

func (t *MathExpressionTestSuite) makeExpression(expr string) (Value, error) {
	t.parser.Flush()

	parseResp := t.parser.ParseString(expr)

	if parseResp.Error != nil {
		return nil, parseResp.Error
	}

	resp, err := t.parser.Exec()

	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}
