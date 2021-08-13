package parser

import (
	"container/list"
)

type ExpressionStack struct {
	list *list.List
	size int
}

type stackItem struct {
	ctx        SystemContext
	expression Expression
}

func newExpressionStack() *ExpressionStack {
	return &ExpressionStack{list: new(list.List)}
}

func (es *ExpressionStack) Size() int {
	return es.size
}

func (es *ExpressionStack) Push(ctx SystemContext, exp Expression) {
	es.list.PushFront(&stackItem{
		ctx:        ctx,
		expression: exp,
	})
	es.size++
}

func (es *ExpressionStack) Pop() (SystemContext, Expression) {
	e := es.list.Front()

	if e == nil {
		return nil, nil
	}

	es.list.Remove(e)
	es.size--
	item := e.Value.(*stackItem)
	return item.ctx, item.expression
}
