package parser

import (
	"fmt"
	"testing"

	"github.com/blkmlk/microshell/internal/models"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"
	"github.com/stretchr/testify/suite"
)

func TestStdExpression(t *testing.T) {
	suite.Run(t, new(stdTestSuite))
}

type stdTestSuite struct {
	suite.Suite
	ctn    di.Container
	parser Parser
	ctx    SystemContext
}

func (t *stdTestSuite) SetupTest() {
	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(
		Definition,
		DefinitionContext,
		DefinitionScope,
		di.Def{
			Name: DefinitionNameCommandTree,
			Build: func(ctn di.Container) (interface{}, error) {
				return List{}, nil
			},
		},
		logger.Definition,
	)
	t.Require().NoError(err)

	ctn := builder.Build()

	t.ctn = ctn
	t.ctx = ctn.Get(DefinitionNameRootScope).(SystemContext)
}

func (t *stdTestSuite) TestStrictBool() {
	checkValue := func(ctx SystemContext, exp Expression, value interface{}) {
		t.Require().Equal(value, exp.Value(ctx).Bool())
	}

	t.Require().NoError(t.testCase(true, `false`, false, checkValue))
	t.Require().NoError(t.testCase(true, `true`, true, checkValue))
	t.Require().NoError(t.testCase(true, `true `, true, checkValue))
	t.Require().NoError(t.testCase(true, `false `, false, checkValue))

	t.Require().NoError(t.testCase(true, `"false"`, false, checkValue))

	t.Require().Error(t.testCase(true, `truee`, nil, nil))
	t.Require().Error(t.testCase(true, `falsee`, nil, nil))
	t.Require().Error(t.testCase(true, `tru `, nil, nil))
	t.Require().Error(t.testCase(true, `fal `, nil, nil))
	t.Require().Error(t.testCase(true, `tru`, nil, nil))
	t.Require().Error(t.testCase(true, `fals`, nil, nil))
	t.Require().Error(t.testCase(true, `hehehe`, nil, nil))
}

func (t *stdTestSuite) TestString() {
	checkValue := func(ctx SystemContext, exp Expression, value interface{}) {
		t.Require().Equal(value, exp.Value(ctx).String())
	}

	t.Require().NoError(t.testCase(false, `"hello"`, "hello", checkValue))
	t.Require().NoError(t.testCase(false, `"hello" `, "hello", checkValue))
	t.Require().NoError(t.testCase(false, `"h"`, "h", checkValue))
	t.Require().NoError(t.testCase(false, `h`, "h", checkValue))

	t.Require().Error(t.testCase(false, `"h`, nil, nil))
	t.Require().Error(t.testCase(false, `h"ello`, nil, nil))
	t.Require().Error(t.testCase(false, `h"ell"o`, nil, nil))
}

func (t *stdTestSuite) TestNumber() {
	checkValue := func(ctx SystemContext, exp Expression, value interface{}) {
		t.Require().Equal(value, exp.Value(ctx).Number())
	}

	t.Require().NoError(t.testCase(false, `"hello"`, 0, checkValue))
	t.Require().NoError(t.testCase(false, `"123"`, 123, checkValue))
	t.Require().NoError(t.testCase(false, `123`, 123, checkValue))
	t.Require().NoError(t.testCase(false, `123 `, 123, checkValue))

	t.Require().Error(t.testCase(false, ` 123`, nil, nil))
}

func (t *stdTestSuite) testCase(strictMode bool, s string, expectedValue interface{}, checkValue func(SystemContext, Expression, interface{})) error {
	exp := NewStdExpression(strictMode)

	var errStr string
	for _, c := range s {
		errStr += string(c)

		resp := exp.Add(t.ctx, models.Rune(c))

		if resp.Err() != nil {
			return fmt.Errorf("error on %s (%v)", errStr, resp.Err())
		}

		if resp.Action() == ResponseGoOut {
			break
		}
	}

	if checkValue != nil {
		checkValue(t.ctx, exp, expectedValue)
	}

	resp := exp.Close(t.ctx)

	if resp.Error != nil {
		return fmt.Errorf("error on close (%v)", resp.Error)
	}

	return nil
}
