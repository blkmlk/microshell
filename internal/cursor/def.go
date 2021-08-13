package cursor

import (
	"github.com/blkmlk/microshell/internal/models"
	"github.com/sarulabs/di/v2"
)

const DefinitionName = "cursor"

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return NewCursor(), nil
		},
	}
)

type Cursor interface {
	NewCursor() Cursor
	String() string
	StringFromPosition() string
	StringFromPositionOffset(offset int) string
	Flush()
	Position() int
	SetPosition(position int)
	GetRune() models.Rune
	Len() int
	Swap() int
	MoveForward() int
	MoveBackward() int
	MoveToPrevWord() int
	MoveToNextWord() int
	MoveToStart() int
	MoveToEnd() int
	DeleteToStart() int
	DeleteToEnd() int
	DeleteToPrevWord() int
	Backspace() int
	Delete() int
	WriteString(str string) (int, error)
	WriteRune(r models.Rune)
	AllWords() []string
	AllWordsFromCursor() []string
	Words() []string
	WordsFromCursor() []string
}
