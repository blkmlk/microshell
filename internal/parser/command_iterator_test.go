package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandIterator_NextOptions(t *testing.T) {
	tree := NewCommandTree()
	tree.Add("abc", &Payload{Value: "abc"})
	tree.Add("afg", &Payload{Value: "afg"})

	it := tree.GetIterator()
	opts := it.NextOptions()
	require.NotNil(t, opts)
	require.Len(t, opts.Options, 2)
	require.Equal(t, "a", opts.Merged)

	tree.Add("aa", &Payload{Value: "aa"})
	it = tree.GetIterator()
	opts = it.NextOptions()
	require.Len(t, opts.Options, 3)

	tree.Add("aabb", &Payload{Value: "aabb"})
	it = tree.GetIterator()
	opts = it.NextOptions()
	require.Len(t, opts.Options, 4)
	require.Equal(t, "a", opts.Merged)

	tree.Add("bo", &Payload{Value: "bo"})
	tree.Add("bodo", &Payload{Value: "bodo"})
	it = tree.GetIterator()
	opts = it.NextOptions()
	require.Len(t, opts.Options, 6)
	require.Empty(t, opts.Merged)

	tree.Use("bo")
	opts = it.NextOptions()
	require.Len(t, opts.Options, 5)

	require.True(t, it.GoNext('b'))
	opts = it.NextOptions()
	require.Len(t, opts.Options, 1)
	require.Equal(t, "odo ", opts.Merged)
}

func TestCommandIterator_NextOptions_Used(t *testing.T) {
	tree := NewCommandTree()
	tree.Add("bo", &Payload{Value: "bo"})
	tree.Add("bodo", &Payload{Value: "bodo"})

	it := tree.GetIterator()
	opts := it.NextOptions()
	require.Len(t, opts.Options, 2)

	tree.Use("bodo")
	opts = it.NextOptions()
	require.Len(t, opts.Options, 1)
	require.Equal(t, "bo", opts.Options[0].Name)
	require.Equal(t, "bo ", opts.Merged)

	tree = NewCommandTree()
	tree.Add("bo", &Payload{Value: "bo"})
	tree.Add("bodo", &Payload{Value: "bodo"})

	it = tree.GetIterator()
	opts = it.NextOptions()
	require.Len(t, opts.Options, 2)

	tree.Use("bo")
	opts = it.NextOptions()
	require.Len(t, opts.Options, 1)
	require.Equal(t, "bodo", opts.Options[0].Name)
	require.Equal(t, "bodo ", opts.Merged)
}

func TestCommandIterator_NextOptions_GoToEnd(t *testing.T) {
	tree := NewCommandTree()
	tree.Add("bo", &Payload{Value: "bo"})
	tree.Add("bodo", &Payload{Value: "bodo"})

	it := tree.GetIterator()
	require.True(t, it.GoToEnd())
	require.NotNil(t, it.current.Payload())
	require.Equal(t, "bo", it.current.Payload().(*Payload).Value)
	require.True(t, it.GoToEnd())
	require.NotNil(t, it.current.Payload())
	require.Equal(t, "bo", it.current.Payload().(*Payload).Value)

	it = tree.GetIterator()
	require.True(t, it.GoNext('b'))
	require.True(t, it.GoNext('o'))
	require.True(t, it.GoToEnd())
	require.Equal(t, "bo", it.current.Payload().(*Payload).Value)

	tree.Use("bo")
	it = tree.GetIterator()
	require.True(t, it.GoNext('b'))
	require.True(t, it.GoNext('o'))
	ops := it.NextOptions()
	require.Len(t, ops.Options, 1)
	require.Equal(t, "do ", ops.Merged)
	require.True(t, it.GoToEnd())
	ops = it.NextOptions()
	require.Len(t, ops.Options, 1)
	require.Equal(t, "bodo", it.current.Payload().(*Payload).Value)

	it = tree.GetIterator()
	ops = it.NextOptions()
	require.Len(t, ops.Options, 1)
	require.Equal(t, "bodo ", ops.Merged)

	it = tree.GetIterator()
	it.GoNext('b')
	it.GoNext('o')
	it.GoNext('d')
	it.GoNext('o')
	ops = it.NextOptions()
	require.Len(t, ops.Options, 1)
	require.Equal(t, " ", ops.Merged)
}

func TestCommandIterator_NextOptions_TwoPaths(t *testing.T) {
	tree := NewCommandTree()
	tree.Add("name", &Payload{Value: "name"})
	tree.Add("value", &Payload{Value: "value"})

	it := tree.GetIterator()
	ops := it.NextOptions()
	require.Len(t, ops.Options, 2)

	tree.Use("value")
	it = tree.GetIterator()
	ops = it.NextOptions()
	require.Len(t, ops.Options, 1)
	require.Equal(t, "name ", ops.Merged)
}
