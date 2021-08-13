package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariableIterator_NextOptions(t *testing.T) {
	tree := NewVariableTree()

	tree.AddGlobal("abc", 123)
	tree.AddGlobal("afg", 123)

	it := tree.GetIterator()
	nextOpts := it.NextOptions()
	require.NotNil(t, nextOpts)
	require.Equal(t, "a", nextOpts.Merged)
	require.Len(t, nextOpts.Options, 2)
	require.Equal(t, "abc", nextOpts.Options[0].Name)
	require.Equal(t, "afg", nextOpts.Options[1].Name)

	tree.AddGlobal("a", 100)
	it = tree.GetIterator()
	nextOpts = it.NextOptions()
	require.NotNil(t, nextOpts)
	require.Equal(t, "a", nextOpts.Merged)
	require.Len(t, nextOpts.Options, 3)
	require.Equal(t, "a", nextOpts.Options[0].Name)
	require.Equal(t, "abc", nextOpts.Options[1].Name)
	require.Equal(t, "afg", nextOpts.Options[2].Name)

	require.True(t, it.Next('a'))
	nextOpts = it.NextOptions()
	require.Empty(t, nextOpts.Merged)
	require.Len(t, nextOpts.Options, 2)
	require.Equal(t, "abc", nextOpts.Options[0].Name)
	require.Equal(t, "afg", nextOpts.Options[1].Name)

	require.True(t, it.Next('f'))
	nextOpts = it.NextOptions()
	require.Equal(t, "g", nextOpts.Merged)
	require.Len(t, nextOpts.Options, 1)
	require.Equal(t, "afg", nextOpts.Options[0].Name)
}
