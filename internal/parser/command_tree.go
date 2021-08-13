package parser

type NextOptions struct {
	AggregatedLevel LevelType
	Options         []*NextOption
	Merged          string
}

type NextOption struct {
	Name  string
	Level LevelType
}

type CommandTree struct {
	root *commandNode
	used *commandNode
}

type Payload struct {
	Level    LevelType
	Value    string
	NextTree *CommandTree
	Payload  interface{}
}

func NewCommandTree() *CommandTree {
	var t CommandTree
	t.root = NewNode()
	t.used = NewNode()

	return &t
}

func (c *CommandTree) Copy() *CommandTree {
	return &CommandTree{
		root: c.root,
		used: NewNode(),
	}
}

func (c *CommandTree) Add(key string, payload *Payload) {
	c.root.Add(key, payload)
}

func (c *CommandTree) GetIterator() *commandIterator {
	return &commandIterator{current: c.root, used: c.used, tree: c}
}

func (c *CommandTree) Use(key string) {
	c.used.Add(key, &Payload{
		Value: key,
	})
}
