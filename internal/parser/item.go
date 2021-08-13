package parser

type LevelType int

const (
	LevelTypeNone     LevelType = 0 << 1
	LevelTypePath     LevelType = 1 << 1
	LevelTypeCommand  LevelType = 2 << 1
	LevelTypeFlag     LevelType = 3 << 1
	LevelTypeOption   LevelType = 4 << 1
	LevelTypeVariable LevelType = 5 << 1
)

type Item struct {
	Level    LevelType
	Payload  interface{}
	Children map[string]*Item
}

func BuildCommandTree(items map[string]*Item) *CommandTree {
	tree := NewCommandTree()

	addItemToTree(tree, items)

	return tree
}

func addItemToTree(tree *CommandTree, items map[string]*Item) {
	for key, item := range items {
		p := &Payload{
			Level:   item.Level,
			Value:   key,
			Payload: item.Payload,
		}

		if len(item.Children) > 0 {
			p.NextTree = NewCommandTree()
			addItemToTree(p.NextTree, item.Children)
		}

		tree.Add(key, p)
	}
}
