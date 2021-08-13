package terminal

import (
	"container/list"

	"github.com/sarulabs/di/v2"
)

const DefinitionNameBuffer = "buffer"

var (
	DefinitionBuffer = di.Def{
		Name: DefinitionNameBuffer,
		Build: func(ctn di.Container) (interface{}, error) {
			return newBuffer(), nil
		},
	}
)

type Buffer interface {
	Push(Output)
	Pop() (Output, bool)
	Len() int
}

type buffer struct {
	stack *list.List
}

func (b *buffer) Len() int {
	return b.stack.Len()
}

func newBuffer() Buffer {
	return &buffer{stack: new(list.List)}
}

func (b *buffer) Push(output Output) {
	b.stack.PushFront(output)
}

func (b *buffer) Pop() (Output, bool) {
	e := b.stack.Back()
	if e == nil {
		return nil, false
	}
	b.stack.Remove(e)
	return e.Value.(Output), true
}
