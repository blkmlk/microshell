package parser

import (
	"sort"
	"strings"

	"github.com/blkmlk/microshell/internal/models"
)

type commandIterator struct {
	tree    *CommandTree
	current *commandNode
	used    *commandNode
}

func (i *commandIterator) IsNext(r models.Rune) bool {
	return i.getNext(r) != nil
}

func (i *commandIterator) GoNext(r models.Rune) bool {
	nextNode := i.getNext(r)

	if nextNode == nil {
		return false
	}

	i.current = nextNode

	return true
}

func (i *commandIterator) getNext(r models.Rune) *commandNode {
	nextNode := i.current.Next(r)

	if nextNode == nil {
		return nil
	}

	if i.used != nil {
		nextUsedNode := i.used.Next(r)

		if nextUsedNode != nil && nextNode.Count()-nextUsedNode.Count() <= 0 {
			return nil
		}

		i.used = nextUsedNode
	}

	return nextNode
}

func (i *commandIterator) GoToEnd() bool {
	var nextCurrentNode, nextUsedNode *commandNode

	current := i.current
	used := i.used

	for {
		if current == nil {
			return false
		}

		if len(current.runes) == 0 {
			break
		}

		if current.Payload() != nil {
			p := current.Payload().(*Payload)
			if _, ok := i.tree.used.nodes[p.Value]; !ok {
				break
			}
		}

		paths := 0
		for r, n := range current.runes {
			if paths > 1 {
				return false
			}

			if used == nil {
				paths++
				nextCurrentNode = n
				continue
			}

			nextUsed := used.Next(r)

			if nextUsed == nil {
				paths++
				nextCurrentNode = n
				continue
			}

			if n.Count()-nextUsed.Count() > 0 {
				paths++
				nextCurrentNode = n
				nextUsedNode = nextUsed
				continue
			}
		}

		if paths != 1 {
			return false
		}

		current = nextCurrentNode
		used = nextUsedNode
	}

	i.current = current

	return true
}

func (i *commandIterator) Level() LevelType {
	if i.current.Payload() == nil {
		return LevelTypeNone
	}

	return i.current.Payload().(*Payload).Level
}

func (i *commandIterator) Value() string {
	if i.current.Payload() == nil {
		return ""
	}

	return i.current.Payload().(*Payload).Value
}

func (i *commandIterator) Payload() interface{} {
	if i.current.Payload() == nil {
		return nil
	}

	return i.current.Payload().(*Payload).Payload
}

func (i *commandIterator) NextTree() *CommandTree {
	if i.current.Payload() == nil {
		return nil
	}

	return i.current.Payload().(*Payload).NextTree
}

func (i *commandIterator) NextOptions() *NextOptions {
	level := LevelTypeNone

	merged, options := findNextOptions(i.tree, i.current)
	for _, o := range options {
		level |= o.Level
	}

	sort.Slice(options, func(i, j int) bool {
		return strings.Compare(options[i].Name, options[j].Name) == -1
	})

	if len(options) == 1 {
		switch level {
		case LevelTypeFlag:
			merged += "="
		default:
			merged += " "
		}
	}

	return &NextOptions{
		AggregatedLevel: level,
		Options:         options,
		Merged:          merged,
	}
}

func findNextOptions(tree *CommandTree, node *commandNode) (string, []*NextOption) {
	var (
		options    []*NextOption
		lastMerged string
		lastRune   models.Rune
	)

	if node.Payload() != nil {
		p := node.Payload().(*Payload)

		if _, ok := tree.used.nodes[p.Value]; !ok {
			options = append(options, &NextOption{
				Name:  p.Value,
				Level: p.Level,
			})
		}
	}

	// TODO: Add a check for used runes in another tree
	unusedRunes := 0
	for r, nextNode := range node.runes {
		merged, opts := findNextOptions(tree, nextNode)

		if len(opts) == 0 {
			continue
		}

		options = append(options, opts...)
		lastMerged = merged
		lastRune = r
		unusedRunes++
	}

	if unusedRunes == 1 {
		return lastRune.String() + lastMerged, options
	}

	return "", options
}
