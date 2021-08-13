package shell

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/blkmlk/microshell/internal/parser"

	"github.com/blkmlk/microshell/internal/cursor"

	"github.com/blkmlk/microshell/internal/logger"

	"github.com/blkmlk/microshell/internal/models"

	"github.com/blkmlk/microshell/internal/history"

	"github.com/blkmlk/microshell/internal/prompt"

	"github.com/sarulabs/di/v2"

	"github.com/blkmlk/microshell/internal/terminal"
)

const (
	KeyCtrlA     = 1
	KeyCtrlB     = 2
	KeyCtrlC     = 3
	KeyCtrlD     = 4
	KeyCtrlE     = 5
	KeyCtrlF     = 6
	KeyCtrlH     = 8
	KeyBackspace = 0x7F
	KeyTab       = 9
	KeyCtrlK     = 11
	KeyCtrlL     = 12
	KeyEnter     = 13
	KeyCtrlN     = 14
	KeyCtrlP     = 16
	KeyCtrlT     = 20
	KeyCtrlU     = 21
	KeyCtrlW     = 23
	KeyEsc       = 27
	KeyUp        = 65
	KeyDown      = 66
	KeyRight     = 67
	KeyLeft      = 68
	KeyB         = 98
	KeyF         = 102
	KeyAltB      = -1
	KeyAltF      = -2
	KeySuggest   = 1000
)

type RenderType string

const (
	RenderTypeFull         RenderType = "full"
	RenderTypeFullTrim     RenderType = "full-trim"
	RenderTypePartial      RenderType = "partial"
	RenderTypePartialClear RenderType = "partial-clear"
	RenderTypeCursorOnly   RenderType = "cursor-only"
	RenderTypeSkip         RenderType = "skip"
)

type ParseType string

const (
	ParseTypeFull ParseType = "parse-full"
	ParseTypeRune ParseType = "parse-rune"
	ParseTypeSkip ParseType = "parse-skip"
)

type Shell struct {
	terminal terminal.Terminal
	history  history.History
	prompt   prompt.Prompt
	parser   parser.Parser
	logger   logger.Logger
	buffer   terminal.Buffer

	renderLevel   int
	usernameColor terminal.Color
	hostnameColor terminal.Color
	colors        map[parser.Object]terminal.Color

	ticker chan bool

	text         []rune
	lines        int
	usedLines    int
	promptOffset int
	oldPrompt    string
	cancel       context.CancelFunc
}

func NewShell(ctn di.Container) *Shell {
	shell := &Shell{
		prompt:        ctn.Get(prompt.DefinitionName).(prompt.Prompt),
		history:       ctn.Get(history.DefinitionName).(history.History),
		terminal:      ctn.Get(terminal.DefinitionName).(terminal.Terminal),
		parser:        ctn.Get(parser.DefinitionName).(parser.Parser),
		logger:        ctn.Get(logger.DefinitionName).(logger.Logger),
		buffer:        ctn.Get(terminal.DefinitionNameBuffer).(terminal.Buffer),
		usernameColor: terminal.ColorBlue,
		hostnameColor: terminal.ColorGreen,
		ticker:        make(chan bool),
		colors:        make(map[parser.Object]terminal.Color),
		lines:         1,
		usedLines:     1,
	}

	shell.prompt.SetHostname("localhost")
	shell.prompt.SetUsername("void")

	return shell
}

func (s *Shell) getCursor() cursor.Cursor {
	return s.history.Cursor()
}

func (s *Shell) getColor(object parser.Object) terminal.Color {
	if color, ok := s.colors[object]; ok {
		return color
	}

	return terminal.ColorWhite
}

func (s *Shell) SetColors(colors map[parser.Object]terminal.Color) {
	s.colors = colors
}

func (s *Shell) render(rType RenderType, offset int) {
	s.logger.WriteMessages(s.getCursor().String(), "--", s.getCursor().StringFromPosition())

	switch rType {
	case RenderTypeFull:
		s.clearSpace()
		s.printPrompt()
		s.printText(true, offset)
		s.updateCursorLocation(0)
	case RenderTypeFullTrim:
		s.trimLines()
		position := s.clearSpace()
		s.printPrompt()
		s.printText(false, offset)
		s.logger.WriteMessages("position", position)
		s.getCursor().SetPosition(position)
		s.updateCursorLocation(0)
	case RenderTypePartial:
		s.printText(false, offset)
		s.updateCursorLocation(0)
	case RenderTypePartialClear:
		s.printText(false, offset)
		s.updateCursorLocation(0)
		s.clearToEnd()
	case RenderTypeCursorOnly:
		s.updateCursorLocation(0)
	case RenderTypeSkip:
	}
}

func (s *Shell) parse(pType ParseType) error {
	s.terminal.HideCursor()
	defer s.terminal.ShowCursor()

	var err error

	switch pType {
	case ParseTypeFull:
		text := s.getCursor().String()
		resp := s.parser.ParseString(text)

		if resp.Error != nil {
			s.logger.WriteMessages("ParseErr:", resp.Error.Error())
			err = resp.Error
		}

		s.colorText(resp.Objects)
		s.updateCursorLocation(0)

		s.logger.WriteMessages("ParseResp:")
		for _, obj := range resp.Objects {
			s.logger.WriteMessages("\tobj: ", obj.Object, ", len:", obj.Length)
		}
	}

	return err
}

func (s *Shell) colorText(objects []*parser.ParsedObject) {
	lastColor := s.terminal.Color()
	c := s.getCursor().NewCursor()
	c.MoveToStart()

	var position = 0
	for _, obj := range objects {
		for i := 0; i < obj.Length; i++ {
			r := c.GetRune()

			position++
			s.terminal.MoveCursorToPosition(s.getCursorLocation(c, -1))
			s.terminal.SetColor(s.getColor(obj.Object))
			s.terminal.WriteToConsole(r.String())
			c.MoveForward()
		}
	}
	s.terminal.SetColor(lastColor)
}

func (s *Shell) clearSpace() int {
	for i := 0; i < s.lines; i++ {
		x := 0
		y := s.terminal.Height() - i

		s.terminal.MoveCursorToPosition(x, y)
		s.terminal.EraseLine()
	}

	s.terminal.MoveCursorToStart()
	return s.getCursor().MoveToStart()
}

func (s *Shell) trimLines() {
	_, y := s.getCursorEndLocation()
	s.lines -= s.terminal.Height() - y
}

func (s *Shell) updateCursorLocation(offset int) {
	s.terminal.MoveCursorToPosition(s.getCursorLocation(s.getCursor(), offset))
}

func (s *Shell) endCursorLocation() {
	s.terminal.MoveCursorToPosition(s.getCursorEndLocation())
}

func (s *Shell) clearToEnd() {
	x, y := s.getCursorEndLocation()

	// TODO: check it later
	if x == 0 {
		x = s.terminal.Width()
		y--
	}

	s.logger.WriteMessages("clear to end x:", x, "y:", y, "len:", s.getCursor().Len())

	s.terminal.MoveCursorToPosition(x, y)
	s.terminal.EraseToEnd()

	for i := y + 1; i <= s.terminal.Height(); i++ {
		s.terminal.MoveCursorToPosition(0, i)
		s.terminal.EraseLine()
	}

	s.terminal.MoveCursorToPosition(s.getCursorLocation(s.getCursor(), 0))
}

func (s *Shell) getCursorLocation(c cursor.Cursor, offset int) (cursorX int, cursorY int) {
	cursorX = s.currentXOffset(c.Position()+offset) + 1
	cursorY = s.currentYOffset(c.Position()+offset) + 1
	return
}

func (s *Shell) getCursorEndLocation() (cursorX int, cursorY int) {
	cursorX = s.currentXOffset(s.getCursor().Len())
	cursorY = s.currentYOffset(s.getCursor().Len()) + 1
	return
}

func (s *Shell) currentXOffset(position int) int {
	// TODO: fix it
	if position < 0 {
		position = -1
	}

	return (position + 1 + s.promptOffset) % s.terminal.Width()
}

func (s *Shell) currentYOffset(position int) int {
	// TODO: fix it
	if position < 0 {
		position = -1
	}

	return (s.terminal.Height() - s.lines) + (position+1+s.promptOffset)/s.terminal.Width()
}

func (s *Shell) printPrompt() {
	var offset int
	s.terminal.SetColor(terminal.ColorWhite)
	offset += s.terminal.WriteToConsole("[")

	s.terminal.SetColor(s.hostnameColor)
	offset += s.terminal.WriteToConsole(s.prompt.Hostname())

	if s.prompt.Username() != "" {
		s.terminal.SetColor(terminal.ColorWhite)
		offset += s.terminal.WriteToConsole("@")

		s.terminal.SetColor(s.usernameColor)
		offset += s.terminal.WriteToConsole(s.prompt.Username())
	}

	s.terminal.SetColor(terminal.ColorWhite)
	offset += s.terminal.WriteToConsole("]")
	offset += s.terminal.WriteToConsole(" " + s.prompt.StartChar())

	s.promptOffset = offset
}

func (s *Shell) printText(full bool, renderOffset int) {
	s.terminal.SetColor(terminal.ColorWhite)

	var text string
	if full {
		text = s.getCursor().String()
	} else {
		text = s.getCursor().StringFromPositionOffset(renderOffset)
		if renderOffset != 0 {
			s.updateCursorLocation(renderOffset - 1)
		}
	}

	s.logger.WriteMessages("text:", text, "pos:", s.getCursor().Position())

	lines := 0
	for text != "" {
		width := s.terminal.Width()
		offset := s.currentXOffset(s.getCursor().Position())
		newLine := false

		s.logger.WriteMessages(offset, width)

		if offset == 0 && s.currentYOffset(s.getCursor().Len()) == s.terminal.Height() {
			newLine = true
		}

		width -= offset

		var t string
		if len(text) <= width || width == 0 {
			t = text
			text = ""
		} else {
			t = text[:width]
			text = text[width:]
		}

		s.terminal.WriteToConsole(t)

		if newLine {
			s.terminal.WriteToConsole("\n")
			lines++
		}
	}

	s.lines += lines
	s.usedLines += lines
	s.logger.WriteMessages("lines", s.lines)
}

func (s *Shell) enter() {
	s.buffer.Push(terminal.NewPlainText("\n"))
	l := s.buffer.Len()
	resp, err := s.parser.Exec()

	if err == nil && resp.Value != nil {
		s.logger.WriteMessages("Resp:", resp.Value.String())
	}
	if s.buffer.Len() > l {
		s.buffer.Push(terminal.NewPlainText("\n"))
	}
}

func (s *Shell) ReadRunes(ctx context.Context) chan models.Rune {
	ch := make(chan models.Rune, 5)
	go func() {
		for {
			rs, err := s.terminal.ReadRunes()

			if err != nil {
				log.Fatal(err)
			}

			var r models.Rune

			if len(rs) == 3 {
				switch rs[2] {
				case KeyUp:
					r = KeyCtrlP
				case KeyDown:
					r = KeyCtrlN
				case KeyRight:
					r = KeyCtrlF
				case KeyLeft:
					r = KeyCtrlB
				}
			} else if len(rs) == 2 && rs[0] == KeyEsc {
				switch rs[1] {
				case KeyB:
					r = KeyAltB
				case KeyF:
					r = KeyAltF
				}
			} else {
				r = models.Rune(rs[0])
			}

			select {
			case <-ctx.Done():
				return
			case ch <- r:
			}
		}
	}()
	return ch
}

func (s *Shell) Run() {
	defer s.terminal.ResetTerminal()

	renderType := RenderTypeFull
	renderOffset := 0
	s.render(renderType, renderOffset)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	var completeCh = make(chan models.Rune)
	ch := s.ReadRunes(ctx)

	var r models.Rune

	for {
		s.logger.WriteMessages("RenderType: ", renderType)
		renderOffset = 0

		select {
		case <-ctx.Done():
			return
		case r = <-ch:
		case r = <-completeCh:
		}

		switch r {
		case KeyTab:
			if s.getCursor().Position() != s.getCursor().Len()-1 {
				continue
			}

			resp := s.parser.Continue()

			if resp == nil {
				continue
			}

			completeCh = make(chan models.Rune, len(resp.Merged))
			for _, c := range resp.Merged {
				completeCh <- models.Rune(c)
			}

			if resp.Merged == "" && len(resp.Options) > 0 {
				s.buffer.Push(terminal.NewPlainText("\n"))

				out := terminal.NewFlexibleTable()
				var opts []string
				for _, opt := range resp.Options {
					opts = append(opts, opt.Option)
					w := terminal.Word{}
					switch opt.Level {
					case parser.LevelTypePath:
						color := s.colors[parser.ObjectPath]
						w.SetColor(color)
					case parser.LevelTypeCommand:
						color := s.colors[parser.ObjectCommand]
						w.SetColor(color)
					case parser.LevelTypeFlag:
						color := s.colors[parser.ObjectMandatoryFlag]
						w.SetColor(color)
					case parser.LevelTypeOption:
						color := s.colors[parser.ObjectOption]
						w.SetColor(color)
					default:
						w.SetColor(terminal.ColorWhite)
					}
					w.SetText(opt.Option)
					out.AddWord(w)
				}
				s.buffer.Push(out)
				s.buffer.Push(terminal.NewPlainText("\n"))

				s.logger.WriteMessages("Options: ", strings.Join(opts, ","))
				s.logger.WriteMessages("Merged: ", resp.Merged)
				completeCh = make(chan models.Rune, 1)
				completeCh <- KeySuggest
				renderType = RenderTypeFullTrim
			}

			continue
		case KeySuggest:
			s.printBuffer()
		case KeyCtrlF:
			s.getCursor().MoveForward()
			renderType = RenderTypeCursorOnly
		case KeyCtrlA:
			s.getCursor().MoveToStart()
			renderType = RenderTypeCursorOnly
		case KeyCtrlB:
			s.getCursor().MoveBackward()
			renderType = RenderTypeCursorOnly
		case KeyCtrlD:
			if s.getCursor().Position() == 0 && s.getCursor().String() == " " {
				s.cancel()
				continue
			}

			s.getCursor().Delete()
			renderType = RenderTypePartialClear
		case KeyCtrlE:
			s.getCursor().MoveToEnd()
			renderType = RenderTypeCursorOnly
		case KeyCtrlH, KeyBackspace:
			s.getCursor().Backspace()
			renderType = RenderTypePartialClear
			renderOffset = -1
		case KeyCtrlU:
			s.getCursor().DeleteToStart()
			renderType = RenderTypePartialClear
		case KeyEnter:
			s.history.Push()
			s.getCursor().Flush()
			s.enter()
			s.printBuffer()
			renderType = RenderTypeFull
		case KeyCtrlK:
			s.getCursor().DeleteToEnd()
			renderType = RenderTypePartialClear
		case KeyCtrlL:
			s.terminal.EraseScreen(2)
			renderType = RenderTypeFullTrim
		case KeyCtrlP:
			if s.history.Prev() {
				s.getCursor().MoveToEnd()
				renderType = RenderTypeFullTrim
			} else {
				renderType = RenderTypeCursorOnly
			}
		case KeyCtrlN:
			if s.history.Next() {
				s.getCursor().MoveToEnd()
				renderType = RenderTypeFullTrim
			} else {
				renderType = RenderTypeCursorOnly
			}
		case KeyCtrlW:
			s.getCursor().DeleteToPrevWord()
			renderType = RenderTypePartialClear
		case KeyCtrlT:
			updated := s.getCursor().Swap()

			if updated == 0 {
				renderType = RenderTypeSkip
			} else {
				renderType = RenderTypePartialClear
				renderOffset = -1
			}
		case KeyAltB:
			s.getCursor().MoveToPrevWord()
			renderType = RenderTypeCursorOnly
		case KeyAltF:
			s.getCursor().MoveToNextWord()
			renderType = RenderTypeCursorOnly
		case KeyCtrlC:
			s.cancel()
		default:
			s.getCursor().WriteRune(r)
			renderType = RenderTypePartial

			s.logger.WriteMessages("char:", int(r))
		}

		s.render(renderType, renderOffset)
		s.parse(ParseTypeFull)
	}
}

func (s *Shell) runParser() {
	go func() {
		timer := time.NewTimer(time.Millisecond * 200)
		for {
			select {
			case <-timer.C:
				s.parse(ParseTypeFull)
			case force := <-s.ticker:
				if force {
					s.parse(ParseTypeFull)
				} else {
					timer = time.NewTimer(time.Millisecond * 100)
				}
			}
		}
	}()
}

func (s *Shell) printBuffer() {
	s.logger.WriteMessages("buffer len", s.buffer.Len())
	if s.buffer.Len() == 0 {
		return
	}

	currentColor := s.terminal.Color()
	for out, exists := s.buffer.Pop(); exists; out, exists = s.buffer.Pop() {
		words := out.Words(s.terminal.Width(), s.terminal.Height())

		for _, w := range words {
			s.terminal.SetColor(w.Color())
			s.terminal.WriteToConsole(w.Text())
		}
	}
	s.terminal.SetColor(currentColor)
}
