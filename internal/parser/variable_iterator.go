package parser

import (
	"sort"
	"strings"

	"github.com/blkmlk/microshell/internal/models"
)

type variableIterator struct {
	currentGlobal *variableNode
	currentLocal  *variableNode
}

func (i *variableIterator) Next(r models.Rune) bool {
	if i.currentGlobal != nil {
		i.currentGlobal = i.currentGlobal.Next(r)
	}

	if i.currentLocal != nil {
		i.currentLocal = i.currentLocal.Next(r)
	}

	return i.currentGlobal != nil || i.currentLocal != nil
}

func (i *variableIterator) GetPayload() interface{} {
	if i.currentLocal != nil && i.currentLocal.variable != nil {
		return i.currentLocal.variable.Payload
	}

	if i.currentGlobal != nil && i.currentGlobal.variable != nil {
		return i.currentGlobal.variable.Payload
	}

	return nil
}

func (i *variableIterator) NextOptions() *NextOptions {
	globalMerged, globalOptions := i.findNextOptions(i.currentGlobal)
	localMerged, localOptions := i.findNextOptions(i.currentLocal)

	options := append([]*NextOption{}, globalOptions...)
	options = append(options, localOptions...)

	merged := ""

	if len(globalMerged) > 0 && len(localMerged) > 0 {
		for i, r := range globalMerged {
			if len(localMerged) > i && rune(localMerged[i]) == r {
				merged += string(r)
			}
		}
	} else if len(globalMerged) == 0 {
		merged = localMerged
	} else {
		merged = globalMerged
	}

	sort.Slice(options, func(i, j int) bool {
		return strings.Compare(options[i].Name, options[j].Name) == -1
	})

	opts := &NextOptions{
		AggregatedLevel: LevelTypeVariable,
		Options:         options,
		Merged:          merged,
	}

	return opts
}

func (i *variableIterator) findNextOptions(node *variableNode) (string, []*NextOption) {
	if node == nil {
		return "", nil
	}

	var (
		options    []*NextOption
		lastRune   models.Rune
		lastMerged string
	)

	for r, n := range node.runes {
		if n.variable != nil {
			options = append(options, &NextOption{Level: LevelTypeVariable, Name: n.variable.Name})
		}

		merged, opts := i.findNextOptions(n)

		options = append(options, opts...)

		lastMerged = merged
		lastRune = r
	}

	if len(node.runes) == 1 {
		return lastRune.String() + lastMerged, options
	}

	return "", options
}
