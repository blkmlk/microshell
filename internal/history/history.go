package history

import (
	"container/list"

	"github.com/blkmlk/microshell/internal/cursor"
)

func newHistory() History {
	h := &history{
		list: new(list.List),
	}
	h.current = h.list.PushBack(&record{
		id:     1,
		cursor: cursor.NewCursor(),
		value:  " ",
	})

	return h
}

type record struct {
	id     int
	value  string
	cursor cursor.Cursor
}

type history struct {
	list    *list.List
	current *list.Element
}

func (h *history) Next() bool {
	if h.current.Next() != nil {
		h.current = h.current.Next()
		return true
	}

	return false
}

func (h *history) Prev() bool {
	if h.current.Prev() != nil {
		h.current = h.current.Prev()
		return true
	}

	return false
}

func (h *history) Value() string {
	return h.currentRecord().value
}

func (h *history) Cursor() cursor.Cursor {
	return h.currentRecord().cursor
}

func (h *history) Load(values []string) {
	h.list = new(list.List)

	i := 1
	for _, v := range values {
		c := cursor.NewCursor()
		c.WriteString(v)

		r := &record{
			id:     i,
			value:  v,
			cursor: c,
		}

		h.current = h.list.PushBack(r)
		i++
	}

	r := new(record)
	r.id = i
	h.current = h.list.PushBack(r)
}

func (h *history) Push() bool {
	value := h.Cursor().String()

	if value == " " {
		return false
	}

	h.currentRecord().cursor = cursor.NewCursorFromString(h.currentRecord().value)

	c := cursor.NewCursorFromString(value)

	front := h.list.Back().Value.(*record)
	front.value = value
	front.cursor = c

	h.current = h.list.PushBack(&record{
		id:     front.id + 1,
		value:  " ",
		cursor: cursor.NewCursor(),
	})

	return true
}

func (h *history) currentRecord() *record {
	return h.current.Value.(*record)
}
