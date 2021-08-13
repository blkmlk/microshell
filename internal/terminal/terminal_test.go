package terminal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTerminal(t *testing.T) {
	terminal, err := newTerminal()
	require.NoError(t, err)

	terminal.WriteToConsole("123")
}
