package parser

import (
	"testing"

	"github.com/blkmlk/microshell/internal/models"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
)

func TestBuildCommandTree(t *testing.T) {
	payload1 := uuid.New().String()
	payload2 := uuid.New().String()
	payload3 := uuid.New().String()
	payload4 := uuid.New().String()

	items := map[string]*Item{
		"path1": {
			Level:   LevelTypePath,
			Payload: payload1,
			Children: map[string]*Item{
				"path2": {
					Level:   LevelTypePath,
					Payload: payload2,
					Children: map[string]*Item{
						"command2": {
							Level:   LevelTypeCommand,
							Payload: payload4,
						},
					},
				},
				"command1": {
					Level:   LevelTypeCommand,
					Payload: payload3,
				},
			},
		},
	}

	tree := BuildCommandTree(items)

	it := tree.GetIterator()

	for _, r := range "path1" {
		require.True(t, it.GoNext(models.Rune(r)))
	}
	require.NotNil(t, it.Payload())
	require.Equal(t, payload1, it.Payload())
	require.True(t, it.GoToEnd())
	require.NotNil(t, it.Payload())
	require.Equal(t, payload1, it.Payload())
	require.Equal(t, LevelTypePath, it.Level())

	oldTree := tree
	tree = it.NextTree()
	require.NotNil(t, tree)
	it = tree.GetIterator()
	require.Len(t, it.NextOptions().Options, 2)
	require.False(t, it.GoToEnd())
	for _, r := range "path2" {
		require.True(t, it.GoNext(models.Rune(r)))
	}
	require.NotNil(t, it.Payload())
	require.Equal(t, payload2, it.Payload())
	require.Equal(t, LevelTypePath, it.Level())

	tree = it.NextTree()
	require.NotNil(t, tree)
	it = tree.GetIterator()
	require.Len(t, it.NextOptions().Options, 1)
	for _, r := range "command2" {
		require.True(t, it.GoNext(models.Rune(r)))
	}
	require.NotNil(t, it.Payload())
	require.Equal(t, payload4, it.Payload())
	require.Equal(t, LevelTypeCommand, it.Level())

	it = oldTree.GetIterator()
	it.GoToEnd()
	tree = it.NextTree()
	it = tree.GetIterator()
	require.Len(t, it.NextOptions().Options, 2)
	for _, r := range "command1" {
		require.True(t, it.GoNext(models.Rune(r)))
	}
	require.NotNil(t, it.Payload())
	require.Equal(t, payload3, it.Payload())
	require.Equal(t, LevelTypeCommand, it.Level())
}
