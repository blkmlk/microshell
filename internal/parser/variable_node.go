package parser

import (
	"github.com/blkmlk/microshell/internal/models"
)

type variable struct {
	Name    string
	Payload interface{}
}

type variableNode struct {
	runes    map[models.Rune]*variableNode
	variable *variable
}

func NewVariableNode() *variableNode {
	return &variableNode{
		runes:    make(map[models.Rune]*variableNode),
		variable: nil,
	}
}

func (n *variableNode) Copy() *variableNode {
	newNode := NewVariableNode()

	for k, v := range n.runes {
		newNode.runes[k] = v
	}

	return newNode
}

func (n *variableNode) Add(key string, payload interface{}) {
	if key == "" {
		return
	}

	newChain := NewVariableNode()
	lastNewNode := newChain
	current := n
	for _, c := range key {
		r := models.Rune(c)

		created := NewVariableNode()
		nextCurrent := current.runes[r]

		if nextCurrent != nil {
			for k, v := range nextCurrent.runes {
				created.runes[k] = v
			}

			if nextCurrent.variable != nil {
				created.variable = nextCurrent.variable
			}

			current = nextCurrent
		}

		lastNewNode.runes[r] = created
		lastNewNode = created
	}

	lastNewNode.variable = &variable{
		Name:    key,
		Payload: payload,
	}

	firstRune := models.Rune(key[0])
	n.runes[firstRune] = newChain.runes[firstRune]
}

func (n *variableNode) Get(key string) *variableNode {
	node := n
	for _, r := range key {
		n := node.runes[models.Rune(r)]

		if n == nil {
			return nil
		}

		node = n
	}

	return node
}

func (n *variableNode) Next(r models.Rune) *variableNode {
	return n.runes[r]
}

// TODO: modify it
func (n *variableNode) FindValues() (string, []*NextOption) {
	var options []*NextOption

	if n.variable != nil {
		options = append(options, &NextOption{Level: LevelTypeVariable, Name: n.variable.Name})
	}

	var lastRune models.Rune
	var lastMerged string
	for r, n := range n.runes {
		merged, values := n.FindValues()

		options = append(options, values...)

		lastMerged = merged
		lastRune = r
	}

	if len(n.runes) == 1 && len(options) == 1 {
		return lastRune.String() + lastMerged, options
	}

	return "", options
}
