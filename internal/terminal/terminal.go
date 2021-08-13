package terminal

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Color int

const (
	ColorBlack Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

type terminal struct {
	in           uintptr
	term         syscall.Termios
	width        int
	height       int
	currentColor Color
}

func newTerminal() (Terminal, error) {
	var t terminal

	t.in = os.Stdin.Fd()

	var st syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, t.in, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&st)), 0, 0, 0); err != 0 {
		return nil, err
	}

	t.term = st

	st.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	st.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, t.in, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&st)), 0, 0, 0); err != 0 {
		return nil, err
	}

	type termInfo struct {
		Height uint16
		Width  uint16
		Xpixel uint16
		Ypixel uint16
	}

	var ti termInfo
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ti))); err != 0 {
		return nil, err
	}
	t.height = int(ti.Height)
	t.width = int(ti.Width)

	return &t, nil
}

func (t *terminal) ResetTerminal() error {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, t.in, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t.term)), 0, 0, 0); err != 0 {
		return err
	}

	return nil
}

func (t *terminal) Width() int {
	return t.width
}

func (t *terminal) Height() int {
	return t.height
}

func (t *terminal) ReadRunes() ([]rune, error) {
	var buf [16]byte
	n, err := syscall.Read(int(t.in), buf[:])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return []rune{}, nil
	}
	if buf[n-1] == '\n' {
		n--
	}
	return []rune(string(buf[:n])), nil
}

func (t *terminal) WriteToConsole(s string) int {
	n, _ := os.Stdout.WriteString(s)
	os.Stdout.Sync()
	return n
}

func (t *terminal) Color() Color {
	return t.currentColor
}

func (t *terminal) SetColor(color Color) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dm", color+30))
	t.currentColor = color
}

func (t *terminal) MoveCursorToPosition(x, y int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%d;%dH", y, x))
}

func (t *terminal) MoveCursorToStart() {
	t.WriteToConsole("\x1b[0G")
}

func (t *terminal) MoveCursorBack(steps int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dD", steps))
}

func (t *terminal) MoveCursorForward(steps int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dC", steps))
}

func (t *terminal) MoveCursorUp(steps int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dA", steps))
}

func (t *terminal) MoveCursorDown(steps int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dB", steps))
}

func (t *terminal) MoveCursorTo(col int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dG", col))
}

func (t *terminal) EraseScreen(mode int) {
	t.WriteToConsole(fmt.Sprintf("\x1b[%dJ", mode))
}

func (t *terminal) ShowCursor() {
	t.WriteToConsole("\x1b[?25h")
}

func (t *terminal) HideCursor() {
	t.WriteToConsole("\x1b[?25l")
}

func (t *terminal) EraseToEnd() {
	t.WriteToConsole("\x1b[0K")
}

func (t *terminal) EraseToStart() {
	t.WriteToConsole("\x1b[1K")
}

func (t *terminal) EraseLine() {
	t.WriteToConsole("\x1b[2K")
}

func (t *terminal) InsertLine() {
	t.WriteToConsole("\x1b[10L")
}

func (t *terminal) ScrollUp() {
	t.WriteToConsole("\x1b[1S")
}

func (t *terminal) ScrollDown() {
	t.WriteToConsole("\x1b[1T")
}
