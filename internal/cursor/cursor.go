package cursor

import (
	"strings"

	"github.com/blkmlk/microshell/internal/models"
)

type cursor struct {
	root     *word
	current  *word
	offset   int
	depth    int
	position int
	length   int
}

type action func() int

func NewCursor() Cursor {
	var w cursor
	w.Flush()

	return &w
}

// TODO: make better
func NewCursorFromString(s string) Cursor {
	var w cursor
	w.Flush()

	if len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}

	w.WriteString(s)

	return &w
}

func (c *cursor) NewCursor() Cursor {
	return &cursor{
		root:     c.root,
		current:  c.root,
		offset:   0,
		depth:    0,
		position: 0,
		length:   1,
	}
}

func (c *cursor) Flush() {
	c.root = newWord()
	c.root.prev = nil

	c.current = c.root
	c.current.SetText(" ")
	c.offset = 0
	c.depth = 0
	c.position = 0
	c.length = 1
}

func (c *cursor) Depth() int {
	return c.depth
}

func (c *cursor) CurrentWord() *word {
	return c.current
}

func (c *cursor) Position() int {
	return c.position
}

func (c *cursor) SetPosition(position int) {
	if position > c.length {
		position = c.length
	}

	c.position = position

	var w *word
	for w = c.root; ; w = w.next {
		if position >= w.Len() {
			position -= w.Len()
		} else {
			c.offset = position
			break
		}
	}

	c.current = w
}

func (c *cursor) GetRune() models.Rune {
	return models.Rune(c.current.Text()[c.offset])
}

func (c *cursor) Len() int {
	return c.length
}

func (c *cursor) MoveForward() int {
	if c.offset < c.current.End() {
		c.offset++
		c.position++
		return 1
	}

	if c.current.next == nil {
		return 0
	}

	c.current = c.current.next
	c.offset = 0
	c.depth++
	c.position++

	return 1
}

func (c *cursor) MoveBackward() int {
	if c.current == c.root {
		if c.offset == 0 {
			c.position = 0
			return 0
		}

		c.offset--
		c.position--
		return 1
	}

	if c.offset == 0 && c.current.prev != nil {
		c.current = c.current.prev
		c.offset = c.current.End()
		c.depth--
		c.position--

		return 1
	}

	c.offset--
	c.position--
	return 1
}

func (c *cursor) MoveToPrevWord() int {
	var n int

	if c.current == c.root {
		if c.current.Len() == 1 {
			return 0
		}

		n = c.offset
		c.offset = 0
		c.position -= n
		return n
	}

	if c.current.prev.IsSpace() {
		n = c.offset + 1
		c.current = c.current.prev
		c.depth--
	} else {
		n += c.offset + 1
		n += c.current.prev.Len()
		c.current = c.current.prev.prev
		c.depth -= 2
	}

	c.offset = c.current.End()
	c.position -= n
	return n
}

func (c *cursor) MoveToNextWord() int {
	var n int
	if c.current.IsSpace() {
		n += c.current.End() - c.offset

		if c.current.next != nil {
			c.current = c.current.next
			c.depth++
			n += c.current.Len()
		}

		c.offset = c.current.End()
		c.position += n
		return n
	}

	if c.offset < c.current.End() {
		n = c.current.End() - c.offset
		c.offset = c.current.End()
		c.position += n
		return n
	}

	// go to space
	if c.current.next != nil {
		c.current = c.current.next
		c.depth++
		n += c.current.Len()
	}

	// go to next text
	if c.current.next != nil {
		c.current = c.current.next
		c.depth++
		n += c.current.Len()
	}

	c.offset = c.current.End()
	c.position += n
	return n
}

func (c *cursor) MoveToStart() int {
	var n int

	for w := c.current.prev; w != nil; w = w.prev {
		n += w.Len()
	}

	n += c.offset

	c.current = c.root
	c.offset = 0
	c.depth = 0
	c.position = 0

	return n
}

func (c *cursor) MoveToEnd() int {
	var n = c.current.End() - c.offset

	for w := c.current.next; w != nil; w = c.current.next {
		n += w.Len()
		c.depth++
		c.current = w
	}

	c.offset = c.current.End()
	c.position += n
	return n
}

func (c *cursor) Swap() int {
	if c.current.next == nil && c.offset == c.current.End() {
		// root
		if c.MoveBackward() == 0 {
			return 0
		}
	}

	if c.current == c.root && c.offset == 0 {
		return 0
	}

	var runes = make([]rune, 2)

	runes[1] = rune(c.current.Text()[c.offset])

	c.Backspace()
	c.MoveForward()

	runes[0] = rune(c.current.Text()[c.offset])

	c.Backspace()
	n, _ := c.WriteString(string(runes))

	return n
}

func (c *cursor) DeleteToStart() int {
	var n int

	if c.current == c.root {
		n += c.offset
	} else {
		n += c.offset + 1
	}

	for w := c.current.prev; w != c.root && w != nil; w = w.prev {
		n += w.Len()
	}

	if c.offset < c.current.End() {
		if c.current != c.root {
			c.root.next = c.current
			c.current.prev = c.root
		}

		c.current.SetText(c.current.Text()[c.offset+1:])
	} else {
		c.root.next = c.current.next

		if c.current.next != nil {
			c.current.next.prev = c.root
		}
	}

	if c.current != c.root {
		c.root.SetText(" ")
	}

	c.current = c.root
	c.offset = 0
	c.depth = 0
	c.position -= n
	c.length -= n
	return n
}

func (c *cursor) DeleteToEnd() int {
	var n = c.current.End() - c.offset

	if c.offset < c.current.End() {
		c.current.SetText(c.current.Text()[:c.offset+1])
	}

	for c := c.current.next; c != nil; c = c.next {
		n += c.Len()
	}

	c.current.next = nil
	c.length -= n
	return n
}

func (c *cursor) DeleteToPrevWord() int {
	var n int
	if c.current == c.root {
		if c.offset == 0 {
			return 0
		}

		n = c.offset
		c.current.SetText(strings.Repeat(" ", c.current.End()-c.offset))
		c.offset = 0
		c.position = 0
		return n
	}

	if c.current.IsSpace() {
		prev := c.current.prev.prev

		n += c.offset + 1
		n += c.current.prev.Len()

		if c.offset < c.current.End() {
			prev.SetText(prev.Text() + c.current.Text()[c.offset+1:])
		}

		c.offset = prev.End()

		if c.current.next != nil {
			c.current.next.prev = prev
		}

		prev.next = c.current.next
		c.current = prev
		c.depth--
		c.position -= n
		c.length -= n

		return n
	}

	if c.offset < c.current.End() {
		c.current.SetText(c.current.Text()[c.offset+1:])
	} else {
		c.current.prev.next = nil

		if c.current.next != nil {
			c.current.prev.next = c.current.next.next
			c.current.prev.SetText(c.current.prev.Text() + c.current.next.Text())
		}
	}

	n += c.offset + 1

	c.current = c.current.prev
	c.offset = c.current.End()
	c.depth--
	c.position -= n
	c.length -= n
	return n
}

func (c *cursor) Backspace() int {
	if c.current == c.root && c.current.Len() == 1 {
		return 0
	}

	n := c.deleteChar()
	return n
}

func (c *cursor) Delete() int {
	if c.offset == c.current.End() && c.current.next == nil {
		return 0
	}

	n := c.MoveForward()
	c.deleteChar()
	return n
}

func (c *cursor) AllWords() []string {
	var words []string

	for w := c.root.next; w != nil; w = w.next {
		words = append(words, w.text)
	}

	return words
}

func (c *cursor) AllWordsFromCursor() []string {
	var words []string

	for w := c.current; w != nil; w = w.next {
		words = append(words, w.text)
	}

	return words
}

func (c *cursor) Words() []string {
	var words []string

	for w := c.root; w != nil; w = w.next {
		if !w.IsSpace() {
			words = append(words, w.text)
		}
	}

	return words
}

func (c *cursor) WordsFromCursor() []string {
	var words []string

	for w := c.current; w != nil; w = w.next {
		if !w.IsSpace() {
			words = append(words, w.text)
		}
	}

	return words
}

func (c *cursor) WriteRune(r models.Rune) {
	if c.current.Len() == 0 {
		c.current.SetText(string(r))
		c.position++
		c.length++
		return
	}

	if (r.IsSpace() && c.current.IsSpace()) ||
		(!r.IsSpace() && !c.current.IsSpace()) {

		if c.offset == c.current.End() {
			c.current.SetText(c.current.Text() + string(r))
		} else {
			c.current.SetText(c.current.Text()[:c.offset+1] + string(r) + c.current.Text()[c.offset+1:])
		}

		c.offset++
		c.position++
		c.length++

		return
	}

	if c.offset == c.current.End() {
		if c.current.next == nil {
			c.current.next = newWord()
			c.current.next.prev = c.current
		}

		c.current.next.SetText(string(r) + c.current.next.Text())

		c.current = c.current.next
		c.offset = 0
		c.depth++
		c.position++
		c.length++

		return
	}

	nextWord := newWord()

	nextWord.SetText(c.current.Text()[c.offset+1:])
	nextWord.next = c.current.next
	c.current.text = c.current.text[:c.offset+1]

	cw := newWord()

	cw.next = nextWord
	nextWord.prev = cw

	cw.prev = c.current
	c.current.next = cw

	c.current = cw
	c.current.SetText(string(r))
	c.offset = 0
	c.position++
	c.length++
}

func (c *cursor) WriteString(str string) (int, error) {
	for _, r := range str {
		c.WriteRune(models.Rune(r))
	}

	return len(str), nil
}

func (c *cursor) String() string {
	builder := new(strings.Builder)

	for c := c.root; c != nil; c = c.next {
		builder.WriteString(c.Text())
	}

	return builder.String()
}

func (c *cursor) StringFromPosition() string {
	return c.StringFromPositionOffset(0)
}

func (c *cursor) StringFromPositionOffset(offset int) string {
	var newOffset = c.offset + offset
	var w = c.current

	// TODO: Simplify it
	for {
		if newOffset > 0 {
			if newOffset > w.End() {
				if w.next == nil {
					newOffset = w.End()
					break
				} else {
					newOffset -= w.End() + 1
					w = w.next
				}
			} else {
				break
			}
		} else if newOffset < 0 {
			if w.prev == nil {
				newOffset = 0
				break
			} else {
				w = w.prev
				newOffset += w.End() + 1
			}
		} else {
			break
		}
	}

	builder := new(strings.Builder)

	builder.WriteString(w.text[newOffset:])

	for c := w.next; c != nil; c = c.next {
		builder.WriteString(c.Text())
	}

	return builder.String()
}

func (c *cursor) deleteChar() int {
	if c.current.Len() == 1 {
		if c.current == c.root {
			return 0
		}

		if c.current.prev != nil {
			c.current.prev.next = c.current.next
		}

		if c.current.next != nil {
			c.current.next.prev = c.current.prev
		}

		c.current = c.current.prev
		c.offset = c.current.End()
		c.depth--
		c.length--
		c.position--

		if c.current.next != nil {
			c.current.SetText(c.current.Text() + c.current.next.Text())
			c.current.next = c.current.next.next
		}

		return 1
	}

	// TODO: check offset decrease
	if c.offset == c.current.End() {
		c.current.SetText(c.current.Text()[:c.offset])
		c.offset = c.current.End()
	} else if c.offset == 0 {
		c.current.SetText(c.current.Text()[c.offset+1:])
		c.current = c.current.prev
		c.offset = c.current.End()
	} else {
		c.current.SetText(c.current.Text()[:c.offset] + c.current.Text()[c.offset+1:])
		c.offset--
	}

	c.length--
	c.position--

	return 1
}
