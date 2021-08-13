package parser

import (
	"testing"

	"github.com/blkmlk/microshell/internal/models"

	"github.com/stretchr/testify/require"
)

func TestVariableTree(t *testing.T) {
	tree := NewVariableTree()

	tree.AddGlobal("abc", NewNumberValue(122))
	tree.AddLocal("ab", NewNumberValue(144))

	v := tree.Get("ab")
	require.Equal(t, NewNumberValue(144), v)

	v = tree.Get("abc")
	require.Equal(t, NewNumberValue(122), v)

	it := tree.GetIterator()
	// local
	require.True(t, it.Next(models.Rune('a')))
	require.True(t, it.Next(models.Rune('b')))
	require.Equal(t, NewNumberValue(144), it.GetPayload())
	// global
	require.True(t, it.Next(models.Rune('c')))
	require.Equal(t, NewNumberValue(122), it.GetPayload())
}
