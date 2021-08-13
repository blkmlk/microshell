package parser

import (
	"fmt"
	"testing"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"

	"github.com/blkmlk/microshell/internal/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/suite"
)

type CommandExpressionTestSuite struct {
	suite.Suite
	parser Parser
	ctx    SystemContext
	root   *CommandTree
	exec   *MockTestExec
}

type expectedValue struct {
	Flags   map[string]string
	Options map[string]bool
}

func (t *CommandExpressionTestSuite) SetupTest() {
	t.exec = new(MockTestExec)

	listDefinition := di.Def{
		Name: DefinitionNameCommandTree,
		Build: func(ctn di.Container) (interface{}, error) {
			return List{Commands: []*Command{
				{
					Path:     []string{"ip", "firewall"},
					Name:     "add",
					Type:     CommandTypeUser,
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
							Mandatory: false,
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

	t.parser = ctn.Get(DefinitionName).(Parser)
	t.ctx = ctn.Get(DefinitionNameRootScope).(SystemContext)
}

func TestNewCommandExpression(t *testing.T) {
	suite.Run(t, new(CommandExpressionTestSuite))
}

func (t *CommandExpressionTestSuite) TestExpression() {
	scope := t.ctx

	// ok
	t.runTest("/;", nil, nil)
	t.runTest("/ip;", nil, nil)
	t.runTest(":ip;", nil, nil)
	t.runTest(": ip f;", nil, nil)
	t.runTest("  /ip;", nil, nil)
	t.runTest("  / ip;", nil, nil)
	t.runTest("/ip firewall;", nil, nil)
	t.runTest("/ip f;", nil, nil)

	// errors
	t.runTest("/:ip;", ErrWrongRune, nil)
	t.runTest(":firewall;", ErrWrongRune, nil)
	t.runTest("/ipf;", ErrWrongRune, nil)
	t.runTest("/ip g;", ErrWrongRune, nil)
	t.runTest("/ip f g;", ErrWrongRune, nil)
	t.runTest("//ip f;", ErrWrongRune, nil)
	t.runTest("/ip / f;", ErrWrongRune, nil)
	t.runTest("/ip firewall add n=1 netdf;", ErrWrongRune, nil)
	t.runTest("/ip firewall add network =11;", ErrWrongRune, nil)
	t.runTest("/ip firewall add network==n1;", ErrWrongRune, nil)

	const (
		networkValue       = "network"
		quotedNetworkValue = "network value"
		areaValue1         = "15"
	)

	var (
		expectedNetwork = &expectedValue{
			Flags: map[string]string{
				"network": networkValue,
			},
			Options: map[string]bool{
				"verbose": false,
			},
		}
		expectedNetworkAndVerbose = &expectedValue{
			Flags: map[string]string{
				"network": networkValue,
			},
			Options: map[string]bool{
				"verbose": true,
			},
		}
		expectedQuotedNetworkAndVerbose = &expectedValue{
			Flags: map[string]string{
				"network": quotedNetworkValue,
			},
			Options: map[string]bool{
				"verbose": true,
			},
		}
		expectedNetworkAreaAndVerbose = &expectedValue{
			Flags: map[string]string{
				"network": networkValue,
				"area":    areaValue1,
			},
			Options: map[string]bool{
				"verbose": true,
			},
		}
	)

	// ok
	t.runTest("/ip firewall add network=n1;", nil, []*expectedValue{{}})
	t.runTest("/ip f a network=n1;", nil, []*expectedValue{{}})
	t.runTest(fmt.Sprintf("/ip firewall add network=\"%s\";", networkValue), nil, []*expectedValue{
		expectedNetwork,
	})
	t.runTest(fmt.Sprintf("/ip firewall add netw=\"%s\";", networkValue), nil, []*expectedValue{
		expectedNetwork,
	})

	t.runTest(fmt.Sprintf("/ip firewall  add  network=\"%s\" verbose;", quotedNetworkValue), nil, []*expectedValue{
		expectedQuotedNetworkAndVerbose,
	})
	t.runTest(fmt.Sprintf("/ip firewall  add  \"%s\" verbose;", quotedNetworkValue), nil, []*expectedValue{
		expectedQuotedNetworkAndVerbose,
	})
	t.runTest(fmt.Sprintf("/ip firewall add %v verbose;", networkValue), nil, []*expectedValue{
		expectedNetworkAndVerbose,
	})
	t.runTest(fmt.Sprintf("/ip firewall add %v verbose;", networkValue), nil, []*expectedValue{
		expectedNetworkAndVerbose,
	})
	t.runTest(fmt.Sprintf("/ip firewall add %v verbose area=%v;", networkValue, areaValue1), nil, []*expectedValue{
		expectedNetworkAreaAndVerbose,
	})
	t.runTest(fmt.Sprintf("/i f a %v verbo a=%v;", networkValue, areaValue1), nil, []*expectedValue{
		expectedNetworkAreaAndVerbose,
	})
	// with variable
	scope.SetGlobalVariable("var", NewStringValue(networkValue))

	t.runTest(fmt.Sprintf("/ip firewall add %v verbose;", "$var"), nil, []*expectedValue{
		expectedNetworkAndVerbose,
	})
	t.runTest(fmt.Sprintf("/ip firewall add network=%v verbose;", "$var"), nil, []*expectedValue{
		expectedNetworkAndVerbose,
	})
	// with command list
	t.runTest(fmt.Sprintf("/ip firewall add [/ip firewall add network=1234] verbose;"), nil, []*expectedValue{
		{
			Flags: map[string]string{
				"network": "1234",
			},
			Options: map[string]bool{
				"verbose": false,
			},
		},
		{
			Flags: map[string]string{
				"network": "",
			},
			Options: map[string]bool{
				"verbose": true,
			},
		},
	})
	t.runTest(fmt.Sprintf("/ip firewall add network=[/ip firewall add network=1234] verbose;"), nil, []*expectedValue{
		{
			Flags: map[string]string{
				"network": "1234",
			},
			Options: map[string]bool{
				"verbose": false,
			},
		},
		{
			Flags: map[string]string{
				"network": "",
			},
			Options: map[string]bool{
				"verbose": true,
			},
		},
	})
	// with math
	t.runTest(fmt.Sprintf("/ip firewall add (\"n\" . 1) verbose;"), nil, []*expectedValue{
		{
			Flags: map[string]string{
				"network": "n1",
			},
			Options: map[string]bool{
				"verbose": true,
			},
		},
	})

	// errors
	t.runTest(fmt.Sprintf("/ip firewall add %v verbose network=n1;", quotedNetworkValue), ErrWrongRune, []*expectedValue{})
	t.runTest(fmt.Sprintf("/ip firewall  add  \"%s\" verbose verb;", networkValue), ErrWrongRune, nil)
}

func (t *CommandExpressionTestSuite) runTest(command string, expectedError error, expectedValues []*expectedValue) {
	invoked := 0

	for _, ev := range expectedValues {
		values := ev
		t.exec.On("Exec", mocks.AnyArgument, mocks.AnyArgument, mocks.AnyArgument).Run(func(args mock.Arguments) {
			flags := args.Get(1).(FlagValues)
			options := args.Get(2).(Options)

			for key, expected := range values.Flags {
				value, _ := flags.Get(key)
				t.Require().Equal(expected, value.String())
			}

			for key, expected := range values.Options {
				value := options.Get(key)
				t.Require().Equal(expected, value)
			}
			invoked++
		}).Return(nil, nil).Once()
	}

	err := t.buildExpression(command)

	if expectedError == nil {
		if err != nil {
			fmt.Println(command)
		}

		t.Require().NoError(err)
	} else {
		if err != expectedError {
			fmt.Println(command)
		}

		t.Require().Error(err)
		t.Require().Equal(expectedError, err)
	}

	t.Require().Equal(len(expectedValues), invoked)
}

func (t *CommandExpressionTestSuite) buildExpression(expr string) error {
	t.parser.Flush()

	resp := t.parser.ParseString(expr)

	if resp.Error != nil {
		return resp.Error
	}

	_, err := t.parser.Exec()

	return err
}
