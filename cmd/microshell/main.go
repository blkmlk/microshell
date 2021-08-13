package main

import (
	"log"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/blkmlk/microshell/internal/parser"

	"github.com/blkmlk/microshell/internal/cursor"
	"github.com/blkmlk/microshell/internal/history"
	"github.com/blkmlk/microshell/internal/prompt"

	"github.com/blkmlk/microshell/internal/shell"
	"github.com/blkmlk/microshell/internal/terminal"
	"github.com/sarulabs/di/v2"
)

func main() {
	builder, err := di.NewBuilder()

	if err != nil {
		log.Fatal(err)
	}

	listDefinition := di.Def{
		Name: parser.DefinitionNameCommandTree,
		Build: func(ctn di.Container) (interface{}, error) {
			return parser.List{Commands: []*parser.Command{
				{
					Type: parser.CommandTypeUser,
					Path: []string{"ip", "firewall"},
					Name: "add",
					Flags: map[string]*parser.Flag{
						"network": {
							Name:      "network",
							Mandatory: true,
							Number:    1,
							ValueType: parser.ValueTypeString,
						},
						"area": {
							Name:      "area",
							Mandatory: true,
							Number:    2,
							ValueType: parser.ValueTypeNumber,
						},
						"ar": {
							Name:      "ar",
							Mandatory: false,
							ValueType: parser.ValueTypeNumber,
						},
						"netlork": {
							Name:      "netlork",
							Mandatory: false,
							ValueType: parser.ValueTypeString,
						},
					},
					Options: map[string]bool{
						"verbose": false,
					},
				},
				{
					Type:           parser.CommandTypeSystem,
					Path:           nil,
					Name:           "global",
					SystemExecFunc: setGlobalVariable,
					OutFunc:        outGlobalVariable,
					Flags: map[string]*parser.Flag{
						"name": {
							Name:      "name",
							Mandatory: true,
							Number:    1,
							ValueType: parser.ValueTypeString,
						},
						"value": {
							Name:      "value",
							Mandatory: true,
							Number:    2,
							ValueType: parser.ValueTypeString,
						},
					},
				},
				{
					Type:           parser.CommandTypeSystem,
					Path:           nil,
					Name:           "local",
					SystemExecFunc: setLocalVariable,
					OutFunc:        outLocalVariable,
					Flags: map[string]*parser.Flag{
						"name": {
							Name:      "name",
							Mandatory: true,
							Number:    1,
							ValueType: parser.ValueTypeString,
						},
						"value": {
							Name:      "value",
							Mandatory: true,
							Number:    2,
							ValueType: parser.ValueTypeString,
						},
					},
				},
				{
					Type:           parser.CommandTypeSystem,
					Path:           nil,
					Name:           "put",
					SystemExecFunc: putValue,
					OutFunc:        nil,
					Flags: map[string]*parser.Flag{
						"value": {
							Name:      "value",
							Mandatory: true,
							Number:    1,
							ValueType: parser.ValueTypeString,
						},
					},
				},
			}}, nil
		},
	}

	err = builder.Add(
		terminal.Definition,
		cursor.Definition,
		prompt.Definition,
		history.Definition,
		logger.Definition,
		terminal.DefinitionBuffer,

		parser.Definition,
		parser.DefinitionScope,
		parser.DefinitionContext,
		listDefinition,
	)

	if err != nil {
		log.Fatal(err)
	}

	ctn := builder.Build()

	sh := shell.NewShell(ctn)
	sh.SetColors(map[parser.Object]terminal.Color{
		parser.ObjectError:             terminal.ColorRed,
		parser.ObjectPath:              terminal.ColorBlue,
		parser.ObjectCommand:           terminal.ColorBlue,
		parser.ObjectOptionalFlag:      terminal.ColorYellow,
		parser.ObjectMandatoryFlag:     terminal.ColorYellow,
		parser.ObjectOption:            terminal.ColorMagenta,
		parser.ObjectValue:             terminal.ColorWhite,
		parser.ObjectEqualSymbol:       terminal.ColorCyan,
		parser.ObjectCurlyBrackets:     terminal.ColorYellow,
		parser.ObjectRoundBrackets:     terminal.ColorYellow,
		parser.ObjectSquareBrackets:    terminal.ColorYellow,
		parser.ObjectOperator:          terminal.ColorYellow,
		parser.ObjectQuotedSymbol:      terminal.ColorCyan,
		parser.ObjectQuotedString:      terminal.ColorCyan,
		parser.ObjectVariableSymbol:    terminal.ColorBlue,
		parser.ObjectVariableName:      terminal.ColorBlue,
		parser.ObjectVariableWrongName: terminal.ColorRed,
	})

	sh.Run()
}

func setGlobalVariable(ctx parser.SystemContext, flags parser.Flags, options parser.Options) (parser.Value, error) {
	name := flags.Get("name").Value(ctx).String()

	e := flags.Get("value").Expression()

	if e.Type() == parser.ExpressionTypeCmdList {
		ctx.SetGlobalVariable(name, e)
	} else {
		ctx.SetGlobalVariable(name, e.Value(ctx))
	}

	ctx.Logger().WriteMessages("Set", name, e.Value(ctx).String())

	return nil, nil
}

func setLocalVariable(ctx parser.SystemContext, flags parser.Flags, options parser.Options) (parser.Value, error) {
	name := flags.Get("name").Value(ctx).String()

	e := flags.Get("value").Expression()

	if e.Type() == parser.ExpressionTypeCmdList {
		ctx.SetLocalVariable(name, e)
	} else {
		ctx.SetLocalVariable(name, e.Value(ctx))
	}

	return nil, nil
}

func outLocalVariable(ctx parser.SystemContext, flags parser.Flags, options parser.Options) {
	nameFlag := flags.Get("name")

	if nameFlag != nil && !ctx.VariableExists(nameFlag.Name) {
		ctx.SetLocalVariable(nameFlag.Value(ctx).String(), parser.NullValue)
	}
}

func outGlobalVariable(ctx parser.SystemContext, flags parser.Flags, options parser.Options) {
	nameFlag := flags.Get("name")

	if nameFlag != nil && !ctx.VariableExists(nameFlag.Name) {
		ctx.SetGlobalVariable(nameFlag.Value(ctx).String(), parser.NullValue)
	}
}

func putValue(ctx parser.SystemContext, flags parser.Flags, options parser.Options) (parser.Value, error) {
	valueFlag := flags.Get("value")
	ctx.Buffer().Push(terminal.NewPlainText(valueFlag.Value(ctx).String()))
	return parser.NullValue, nil
}
