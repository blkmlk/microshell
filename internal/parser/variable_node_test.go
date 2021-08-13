package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariableNode(t *testing.T) {
	original := NewVariableNode()
	original.Add("a", 1)
	original.Add("abc", 123)

	require.Equal(t, 1, original.Get("a").variable.Payload.(int))
	require.Equal(t, 123, original.Get("abc").variable.Payload.(int))

	copied := original.Copy()
	copied.Add("a", 5)
	copied.Add("abcd", 10)

	require.Equal(t, 1, original.Get("a").variable.Payload.(int))
	require.Equal(t, 123, original.Get("abc").variable.Payload.(int))
	require.Nil(t, original.Get("abcd"))

	require.Equal(t, 5, copied.Get("a").variable.Payload.(int))
	require.Equal(t, 123, copied.Get("abc").variable.Payload.(int))
	require.Equal(t, 10, copied.Get("abcd").variable.Payload.(int))

	copied2 := copied.Copy()
	copied2.Add("abcdf", 20)

	require.Equal(t, 5, copied2.Get("a").variable.Payload.(int))
	require.Equal(t, 123, copied2.Get("abc").variable.Payload.(int))
	require.Equal(t, 10, copied2.Get("abcd").variable.Payload.(int))
	require.Equal(t, 20, copied2.Get("abcdf").variable.Payload.(int))
}
