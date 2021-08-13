package parser

import (
	"testing"

	"github.com/blkmlk/microshell/internal/mocks"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/suite"
)

func TestVariable(t *testing.T) {
	suite.Run(t, new(variableTestSuite))
}

type testExpression struct {
	ReturnValue string
	Vars        map[string]string
}

type variableTestSuite struct {
	suite.Suite
	ctn     di.Container
	parser  Parser
	ctx     SystemContext
	mockExp *MockExpression
}

func execGetVar(ctx Context, flags FlagValues, options Options) (Value, error) {
	v, _ := flags.Get("value")
	return v, nil
}

func (t *variableTestSuite) SetupTest() {
	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(
		Definition,
		DefinitionContext,
		DefinitionScope,
		logger.Definition,
		di.Def{
			Name: DefinitionNameCommandTree,
			Build: func(ctn di.Container) (interface{}, error) {
				return List{Commands: []*Command{
					{
						Path:     nil,
						Name:     "get",
						Type:     CommandTypeUser,
						ExecFunc: execGetVar,
						Flags: map[string]*Flag{
							"value": {
								Name:      "value",
								Mandatory: true,
								Number:    1,
								ValueType: ValueTypeString,
							},
						},
					},
				}}, nil
			},
		},
	)
	t.Require().NoError(err)

	ctn := builder.Build()

	t.ctn = ctn
	t.parser = ctn.Get(DefinitionName).(Parser)
	t.ctx = ctn.Get(DefinitionNameRootScope).(SystemContext)

	t.mockExp = new(MockExpression)
}

func (t *variableTestSuite) TestVariable() {
	var value Value

	ctx := t.ctx

	ctx.SetGlobalVariable("test", NewNumberValue(123))
	value = t.buildExpression("$test", nil, nil)
	t.Require().Equal("123", value.String())

	value = t.buildExpression("$tes", nil, nil)
	t.Require().Equal("", value.String())

	ctx.SetGlobalVariable("exp", t.mockExp)
	value = t.buildExpression("$exp var1=value1", nil, &testExpression{
		ReturnValue: "123",
		Vars: map[string]string{
			"var1": "value1",
		},
	})
	t.Require().Equal("123", value.String())

	ctx.SetGlobalVariable("exp", t.mockExp)
	value = t.buildExpression("$exp var1=value1", nil, &testExpression{
		ReturnValue: "123",
		Vars: map[string]string{
			"var1": "value1",
		},
	})
	t.Require().Equal("123", value.String())

	value = t.buildExpression("$exp   v=value", nil, &testExpression{
		ReturnValue: "123",
		Vars: map[string]string{
			"v": "value",
		},
	})
	t.Require().Equal("123", value.String())

	value = t.buildExpression("$exp v=[/get value=hh; /get value=hello]", nil, &testExpression{
		ReturnValue: "123",
		Vars: map[string]string{
			"v": "hello",
		},
	})
	t.Require().Equal("123", value.String())

	// with math
	value = t.buildExpression("$exp v=(1 . 2)", nil, &testExpression{
		ReturnValue: "123",
		Vars: map[string]string{
			"v": "12",
		},
	})
	t.Require().Equal("123", value.String())

	// errors
	value = t.buildExpression("$tes v 123", ErrWrongRune, nil)
	t.Require().Equal("", value.String())

	value = t.buildExpression("$tes v =123", ErrWrongRune, nil)
	t.Require().Equal("", value.String())

	value = t.buildExpression("$tes =123", ErrWrongRune, nil)
	t.Require().Equal("", value.String())

	value = t.buildExpression("$tes name==123", ErrWrongRune, nil)
	t.Require().Equal("", value.String())

	value = t.buildExpression("$$tes", ErrWrongRune, nil)
	t.Require().Equal("", value.String())

	value = t.buildExpression("tes", ErrWrongRune, nil)
	t.Require().Equal("", value.String())
}

func (t *variableTestSuite) buildExpression(s string, expectedError error, te *testExpression) Value {
	invoked := false
	if te != nil {
		t.mockExp.On("Value", mocks.AnyArgument).Run(func(args mock.Arguments) {
			ctx := args.Get(0).(SystemContext)
			invoked = true
			for name, value := range te.Vars {
				t.Require().Equal(value, ctx.GetVariable(name).String())
			}
		}).Return(NewStringValue(te.ReturnValue)).Once()
	}

	t.parser.Flush()

	parseResp := t.parser.ParseString(s)

	if parseResp.Error != nil {
		if expectedError != nil {
			t.Require().Error(parseResp.Error)
			t.Require().Equal(expectedError, parseResp.Error)
		} else {
			t.Fail(parseResp.Error.Error())
		}
	}

	resp, err := t.parser.Exec()

	if err != nil {
		if expectedError != nil {
			t.Require().Error(err)
			t.Require().Equal(expectedError, err)
		} else {
			t.Fail(err.Error())
		}

		return NullValue
	}

	if te != nil {
		t.Require().True(invoked)
	}

	return resp.Value
}
