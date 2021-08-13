package parser

import (
	"context"

	"github.com/blkmlk/microshell/internal/terminal"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/sarulabs/di/v2"
)

type Context interface {
	context.Context
	Buffer() terminal.Buffer
}

type SystemContext interface {
	Context
	CommandTree() *CommandTree
	SetCommandTree(commandTree *CommandTree)
	CommandRoot() *CommandTree
	SetCommandRoot(commandRoot *CommandTree)
	VariableTree() *VariableTree
	GetVariablePayload(name string) interface{}
	VariableExists(name string) bool
	GetVariable(name string) Value
	SetGlobalVariable(name string, value interface{})
	SetLocalVariable(name string, value interface{})
	Ctx() context.Context
	WithContext(ctx context.Context) SystemContext
	New() SystemContext
	Copy() SystemContext
	Logger() logger.Logger
}

type systemContext struct {
	context.Context
	commandTree  *CommandTree
	commandRoot  *CommandTree
	variableTree *VariableTree
	logger       logger.Logger
	buffer       terminal.Buffer
}

func newRootContext(ctn di.Container) (SystemContext, error) {
	scope := &systemContext{
		Context:      ctn.Get(DefinitionNameContext).(context.Context),
		logger:       ctn.Get(logger.DefinitionName).(logger.Logger),
		buffer:       ctn.Get(terminal.DefinitionNameBuffer).(terminal.Buffer),
		variableTree: NewVariableTree(),
	}

	list := ctn.Get(DefinitionNameCommandTree).(List)

	items, err := list.Items()

	if err != nil {
		return nil, err
	}

	scope.SetCommandTree(BuildCommandTree(items))
	scope.SetCommandRoot(scope.CommandTree())

	return scope, nil
}

func (p *systemContext) Buffer() terminal.Buffer {
	return p.buffer
}

func (p *systemContext) Logger() logger.Logger {
	return p.logger
}

func (p *systemContext) WithContext(ctx context.Context) SystemContext {
	p.Context = ctx
	return p
}

func (p *systemContext) Ctx() context.Context {
	return p.Context
}

func (p *systemContext) VariableTree() *VariableTree {
	return p.variableTree
}

func (p *systemContext) CommandRoot() *CommandTree {
	return p.commandRoot
}

func (p *systemContext) SetCommandRoot(commandRoot *CommandTree) {
	p.commandRoot = commandRoot
}

func (p *systemContext) SetCommandTree(commandTree *CommandTree) {
	p.commandTree = commandTree
}

func (p *systemContext) GetVariablePayload(name string) interface{} {
	return p.variableTree.Get(name)
}

func (p *systemContext) New() SystemContext {
	return &systemContext{
		Context:      p.Context,
		commandTree:  p.commandTree,
		commandRoot:  p.commandTree,
		variableTree: p.variableTree.Copy(),
		logger:       p.logger,
		buffer:       p.buffer,
	}
}

func (p *systemContext) Copy() SystemContext {
	return &systemContext{
		Context:      p.Context,
		commandTree:  p.commandTree,
		commandRoot:  p.commandRoot,
		variableTree: p.variableTree,
		logger:       p.logger,
		buffer:       p.buffer,
	}
}

func (p *systemContext) CommandTree() *CommandTree {
	return p.commandTree
}

func (p *systemContext) SetGlobalVariable(name string, value interface{}) {
	p.variableTree.AddGlobal(name, value)
}

func (p *systemContext) SetLocalVariable(name string, value interface{}) {
	p.variableTree.AddLocal(name, value)
}

func (p *systemContext) GetVariable(name string) Value {
	payload := p.variableTree.Get(name)

	if payload == nil {
		return NullValue
	}

	if valuer, ok := payload.(Valuer); ok {
		return valuer.Value(p)
	}

	return payload.(Value)
}

func (p *systemContext) VariableExists(name string) bool {
	v := p.variableTree.get(name) != nil
	return v
}
