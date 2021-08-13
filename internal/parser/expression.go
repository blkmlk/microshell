package parser

import "github.com/blkmlk/microshell/internal/models"

type ExpressionType string

const (
	ExpressionTypeCmd     = "expression-cmd"
	ExpressionTypeCmdList = "expression-cmd-list"
	ExpressionTypeMath    = "expression-math"
	ExpressionTypeVar     = "expression-var"
	ExpressionTypeStd     = "expression-std"
)

type Expression interface {
	Valuer
	Type() ExpressionType
	Add(ctx SystemContext, r models.Rune) *Response
	Complete(ctx SystemContext) *CompleteResponse
	Close(ctx SystemContext) *CloseResponse
}

type CompleteResponse struct {
	Options []*CompleteOption
	Merged  string
}

type CompleteOption struct {
	Level  LevelType
	Option string
}
