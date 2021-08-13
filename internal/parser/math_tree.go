package parser

import (
	"errors"
)

var (
	ErrWrongType       = errors.New("wrong type")
	ErrWrongOperator   = errors.New("wrong operator")
	ErrNoMandatoryFlag = errors.New("no mandatory flag")
)

const (
	OperatorTypeCompare     Operator = 1 << 3
	OperatorTypeConcatenate Operator = 1 << 4
	OperatorTypeAddition    Operator = 1 << 5
	OperatorTypeMultiply    Operator = 1 << 6
	OperatorTypeUnary       Operator = 1 << 7
	OperatorTypeMask        Operator = 0xf8 // 1111 1000
)

const (
	OperatorEqual          = OperatorTypeCompare
	OperatorNotEqual       = OperatorTypeCompare + 1
	OperatorGreater        = OperatorTypeCompare + 2
	OperatorGreaterOrEqual = OperatorTypeCompare + 3
	OperatorLess           = OperatorTypeCompare + 4
	OperatorLessOrEqual    = OperatorTypeCompare + 5
	OperatorConcatenate    = OperatorTypeConcatenate
	OperatorPlus           = OperatorTypeAddition
	OperatorMinus          = OperatorTypeAddition + 1
	OperatorMultiply       = OperatorTypeMultiply
	OperatorDivide         = OperatorTypeMultiply + 1
	OperatorNot            = OperatorTypeUnary
)

type (
	Operator int
)

func (o Operator) LessOrEqualThan(op Operator) bool {
	return o&OperatorTypeMask <= op&OperatorTypeMask
}

func (o Operator) LessThan(op Operator) bool {
	return o&OperatorTypeMask < op&OperatorTypeMask
}

func (o Operator) Equals(op Operator) bool {
	return o&OperatorTypeMask == op&OperatorTypeMask
}

func (o Operator) IsType(t Operator) bool {
	return o&OperatorTypeMask == t
}

// CommandTree
type mathTree struct {
	root *mathNode
}

func NewMathTree() *mathTree {
	return &mathTree{
		root: new(mathNode),
	}
}

func (mt *mathTree) Add(item interface{}) {
	mt.root = mt.root.Add(item)
}

func (mt *mathTree) Value(ctx SystemContext) (Value, error) {
	return mt.root.Value(ctx)
}
