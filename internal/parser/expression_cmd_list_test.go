package parser

import (
	"fmt"
	"testing"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/blkmlk/microshell/internal/mocks"
	"github.com/blkmlk/microshell/internal/models"
	"github.com/stretchr/testify/mock"

	"github.com/sarulabs/di/v2"
	"github.com/stretchr/testify/suite"
)

type CommandListExpressionTestSuite struct {
	suite.Suite
	ctn    di.Container
	parser Parser
	ctx    SystemContext
	root   *CommandTree
	exec   *MockTestExec
}

func execSetVar(ctx SystemContext, flags Flags, options Options) (Value, error) {
	name := flags.Get("name").Value(ctx)
	value := flags.Get("value")

	//fmt.Println("set", name, value)
	ctx.SetLocalVariable(name.String(), value.Value(ctx))

	return NullValue, nil
}

func (t *CommandListExpressionTestSuite) SetupTest() {
	t.exec = new(MockTestExec)

	listDefinition := di.Def{
		Name: DefinitionNameCommandTree,
		Build: func(ctn di.Container) (interface{}, error) {
			return List{Commands: []*Command{
				{
					Type:     CommandTypeUser,
					Path:     []string{"ip", "firewall"},
					Name:     "add",
					ExecFunc: t.exec.Exec,
					Flags: map[string]*Flag{
						"network": {
							Name:      "network",
							Mandatory: true,
							Number:    1,
							ValueType: ValueTypeString,
						},
						"area": {
							Name:      "area",
							Mandatory: true,
							Number:    2,
							ValueType: ValueTypeNumber,
						},
						"netlork": {
							Name:      "netlork",
							Mandatory: false,
							ValueType: ValueTypeString,
						},
					},
					Options: map[string]bool{
						"verbose": false,
					},
				},
				{
					Type:           CommandTypeSystem,
					Path:           nil,
					Name:           "set",
					SystemExecFunc: execSetVar,
					Flags: map[string]*Flag{
						"name": {
							Name:      "name",
							Mandatory: true,
							Number:    1,
							ValueType: ValueTypeString,
						},
						"value": {
							Name:      "value",
							Mandatory: true,
							Number:    2,
							ValueType: ValueTypeString,
						},
					},
				},
			}}, nil
		},
	}

	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(
		Definition,
		DefinitionContext,
		DefinitionScope,
		logger.Definition,
		listDefinition,
	)
	t.Require().NoError(err)

	ctn := builder.Build()

	t.ctn = ctn
	t.parser = ctn.Get(DefinitionName).(Parser)
	t.ctx = ctn.Get(DefinitionNameRootScope).(SystemContext)
}

func TestCommandListExpression(t *testing.T) {
	suite.Run(t, new(CommandListExpressionTestSuite))
}

func (t *CommandListExpressionTestSuite) TestCommandList() {
	// errors
	t.runTest("]", ErrWrongRune, 0)
	t.runTest("[]]", ErrWrongRune, 0)
	t.runTest("[//]", ErrWrongRune, 0)
	t.runTest(")", ErrWrongRune, 0)
	t.runTest("[)", ErrWrongRune, 0)
	t.runTest("()", ErrNotFinished, 0)

	// ok
	t.runTest("[]", nil, 0)
	t.runTest("[/]", nil, 0)
	t.runTest("[/;/;/;]", nil, 0)
	t.runTest("[/;/;/]", nil, 0)
	t.runTest("([])", nil, 0)
	t.runTest("[ /ip firewall add network=n1 area=a1]", nil, 1)
	t.runTest("[/ip firewall add network=n1 area=a1; (1 + 2)]", nil, 1)
	t.runTest("[/ip firewall add network=n1 area=a1; (1 + 2);;;]", nil, 1)
	t.runTest(";;[];", nil, 0)
	t.runTest("[{}]", nil, 0)

	// scopes
	varTest := new(MockExpression)
	varTest.On("Value", mocks.AnyArgument).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(SystemContext)
			v := ctx.GetVariable("var").String()
			t.Require().Equal("123", v)
		}).Return(NullValue, nil).Once()
	t.ctx.SetGlobalVariable("test", varTest)
	t.runTest("[/set ab 123];[$test var=$ab]", nil, 0)

	// with math
	t.runTest("(1 + 2);;", nil, 0)

	// variables
	t.runTest("[ $hello]", nil, 0)

	f := new(MockExpression)
	invoked := false
	f.On("Value", mocks.AnyArgument).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(SystemContext)
			v := ctx.GetVariable("var2").String()
			t.Require().Equal("123", v)
			invoked = true
		}).Return(NullValue).Once()

	t.ctx.SetGlobalVariable("hello", f)

	t.runTest("[$hello var2=123]", nil, 0)
	t.Require().True(invoked)

	t.runTest("{}", nil, 0)

	invoked = false
	f.On("Value", mocks.AnyArgument).
		Run(func(args mock.Arguments) {
			invoked = true
		}).Return(NullValue).Once()
	t.ctx.SetGlobalVariable("hello", f)
	t.runTest("{$hello}", nil, 0)
	t.Require().True(invoked)

	invoked = false
	f.On("Value", mocks.AnyArgument).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(SystemContext)
			v := ctx.GetVariable("f").String()
			t.Require().Equal("10", v)
			invoked = true
		}).Return(NullValue).Once()
	t.ctx.SetGlobalVariable("hello", f)
	t.runTest("{/set var3 10; $hello f=$var3}", nil, 0)
	t.Require().True(invoked)

	invoked = false
	f.On("Value", mocks.AnyArgument).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(SystemContext)
			v := ctx.GetVariable("f").String()
			t.Require().Equal("", v)
			invoked = true
		}).Return(NullValue).Once()
	t.ctx.SetGlobalVariable("hello", f)
	t.runTest("{/set var4 10};[$hello f=$var4]", nil, 0)
	t.Require().True(invoked)
}

func (t *CommandListExpressionTestSuite) runTest(command string, expectedError error, count int) {
	invoked := 0

	for i := 0; i < count; i++ {
		t.exec.On("Exec", mocks.AnyArgument, mocks.AnyArgument, mocks.AnyArgument).Run(func(args mock.Arguments) {
			invoked++
		}).Return(nil, nil).Once()
	}

	err := t.buildExpression(command)

	if expectedError == nil {
		t.Require().NoError(err)
	} else {
		t.Require().Error(err)
		t.Require().Equal(expectedError, err)
	}

	t.Require().Equal(count, invoked)
}

func (t *CommandListExpressionTestSuite) buildExpression(expr string) error {
	t.parser.Flush()
	var err error
	for i, r := range expr {
		fmt.Println(expr[0 : i+1])
		_, err = t.parser.Add(models.Rune(r))

		if err != nil {
			return err
		}
	}

	_, err = t.parser.Exec()

	return err
}
