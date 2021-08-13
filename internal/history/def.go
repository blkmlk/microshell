package history

import (
	"github.com/blkmlk/microshell/internal/cursor"
	"github.com/sarulabs/di/v2"
)

const DefinitionName = "history"

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return newHistory(), nil
		},
	}
)

type History interface {
	Load(values []string)
	Next() bool
	Prev() bool
	Cursor() cursor.Cursor
	Value() string
	Push() bool
}
