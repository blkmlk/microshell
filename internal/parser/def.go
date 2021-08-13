package parser

import (
	"context"

	"github.com/blkmlk/microshell/internal/models"
	"github.com/sarulabs/di/v2"
)

const (
	DefinitionName            = "parser"
	DefinitionNameContext     = "global-context"
	DefinitionNameRootScope   = "parser-global-scope"
	DefinitionNameCommandTree = "parser-command-tree"
)

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return newParser(ctn), nil
		},
	}
	DefinitionScope = di.Def{
		Name: DefinitionNameRootScope,
		Build: func(ctn di.Container) (interface{}, error) {
			return newRootContext(ctn)
		},
	}
	DefinitionContext = di.Def{
		Name: DefinitionNameContext,
		Build: func(ctn di.Container) (interface{}, error) {
			return context.Background(), nil
		},
	}
	DefinitionCommandTree = di.Def{
		Name: DefinitionNameCommandTree,
		Build: func(ctn di.Container) (interface{}, error) {
			return List{}, nil
		},
	}
)

type Parser interface {
	Flush()
	IsFlushed() bool
	Add(r models.Rune) (*ParseRuneResponse, error)
	ParseString(s string) *ParseStringResponse
	Exec() (*ExecResponse, error)
	Continue() *CompleteResponse
}
