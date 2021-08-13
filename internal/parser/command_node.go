package parser

import (
	"github.com/blkmlk/microshell/internal/models"
)

type commandNode struct {
	runes   map[models.Rune]*commandNode
	nodes   map[string]*commandNode
	payload interface{}
	count   int
}

func NewNode() *commandNode {
	return &commandNode{
		runes: make(map[models.Rune]*commandNode),
		nodes: make(map[string]*commandNode),
	}
}

func (n *commandNode) Add(key string, payload interface{}) {
	node := n
	for _, r := range key {
		child, ok := node.runes[models.Rune(r)]

		if !ok {
			child = NewNode()
			node.runes[models.Rune(r)] = child
		}

		child.count++

		node = child
	}

	node.payload = payload
	n.nodes[key] = node
}

func (n *commandNode) Use(key string) {
	node := n
	for _, r := range key {
		child, ok := node.runes[models.Rune(r)]
		if !ok {
			return
		}
		if child.count > 0 {
			child.count--
		}
		node = child
	}
}

func (n *commandNode) Next(r models.Rune) *commandNode {
	return n.runes[r]
}

func (n *commandNode) Payload() interface{} {
	return n.payload
}

func (n *commandNode) Count() int {
	return n.count
}
