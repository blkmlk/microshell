package parser

type VariableTree struct {
	global *variableNode
	local  *variableNode
}

func NewVariableTree() *VariableTree {
	return &VariableTree{
		global: NewVariableNode(),
		local:  NewVariableNode(),
	}
}

func (t *VariableTree) Copy() *VariableTree {
	return &VariableTree{
		global: t.global,
		local:  t.local.Copy(),
	}
}

func (t *VariableTree) AddGlobal(name string, value interface{}) {
	t.global.Add(name, value)
}

func (t *VariableTree) AddLocal(name string, value interface{}) {
	t.local.Add(name, value)
}

func (t *VariableTree) Get(name string) interface{} {
	v := t.get(name)

	if v == nil {
		return NullValue
	}

	return v
}

func (t *VariableTree) get(name string) interface{} {
	n := t.local.Get(name)
	if n != nil && n.variable != nil {
		return n.variable.Payload
	}

	n = t.global.Get(name)
	if n != nil && n.variable != nil {
		return n.variable.Payload
	}

	return nil
}

func (t *VariableTree) GetIterator() *variableIterator {
	return &variableIterator{
		currentGlobal: t.global,
		currentLocal:  t.local,
	}
}
