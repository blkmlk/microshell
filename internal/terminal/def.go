package terminal

import (
	"github.com/sarulabs/di/v2"
)

const (
	DefinitionName = "terminal"
)

var (
	Definition = di.Def{
		Name: DefinitionName,
		Build: func(ctn di.Container) (interface{}, error) {
			return newTerminal()
		},
	}
)

type Terminal interface {
	ResetTerminal() error
	Width() int
	Height() int
	ReadRunes() ([]rune, error)
	WriteToConsole(s string) int
	Color() Color
	SetColor(color Color)
	MoveCursorToPosition(x, y int)
	MoveCursorToStart()
	MoveCursorBack(steps int)
	MoveCursorForward(steps int)
	MoveCursorUp(steps int)
	MoveCursorDown(steps int)
	MoveCursorTo(col int)
	EraseScreen(mode int)
	ShowCursor()
	HideCursor()
	EraseToEnd()
	EraseToStart()
	EraseLine()
	InsertLine()
	ScrollUp()
	ScrollDown()
}
