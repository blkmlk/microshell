package terminal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuffer(t *testing.T) {
	b := newBuffer()

	require.Equal(t, 0, b.Len())

	out := NewFlexibleTable()
	b.Push(out)
	require.Equal(t, 1, b.Len())

	e, ok := b.Pop()
	require.True(t, ok)
	require.Equal(t, out, e)

	e, ok = b.Pop()
	require.False(t, ok)
	require.Nil(t, e)
}
