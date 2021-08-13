package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	tree := NewCommandTree()

	tree.Add("abc", &Payload{
		Level: LevelTypePath,
		Value: "abc",
	})

	tree.Add("abg", &Payload{
		Level: LevelTypePath,
		Value: "abg",
	})

	it := tree.GetIterator()
	require.True(t, it.GoNext('a'))
	require.True(t, it.GoNext('b'))
	require.True(t, it.GoNext('c'))

	it = tree.GetIterator()
	require.True(t, it.GoNext('a'))
	require.True(t, it.GoNext('b'))
	require.True(t, it.GoNext('g'))

	it = tree.GetIterator()

	opts := it.NextOptions().Options
	require.Len(t, opts, 2)

	it = tree.GetIterator()
	require.False(t, it.GoToEnd())

	tree.Use("abg")
	it = tree.GetIterator()

	opts = it.NextOptions().Options
	require.Len(t, opts, 1)

	require.True(t, it.GoNext('a'))
	require.True(t, it.GoNext('b'))
	require.False(t, it.GoNext('g'))
	require.True(t, it.GoNext('c'))

	it = tree.GetIterator()
	require.Equal(t, LevelTypeNone, it.Level())
	require.True(t, it.GoToEnd())
	require.True(t, it.GoToEnd())
	require.Equal(t, LevelTypePath, it.Level())
}
