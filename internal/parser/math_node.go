package parser

import "errors"

type mathNode struct {
	Item  interface{}
	Left  *mathNode
	Right *mathNode
}

func (m *mathNode) Value(ctx SystemContext) (Value, error) {
	switch i := m.Item.(type) {
	case Valuer:
		return i.Value(ctx), nil
	case Operator:
		var left Value

		if m.Left != nil {
			var err error
			left, err = m.Left.Value(ctx)

			if err != nil {
				return nil, err
			}
		}

		right, err := m.Right.Value(ctx)

		if err != nil {
			return nil, err
		}

		return m.processOperator(i, left, right)
	case nil:
		return nil, nil
	default:
		return nil, errors.New("unknown item type")
	}
}

func (m *mathNode) processOperator(op Operator, left, right Value) (Value, error) {
	if op == OperatorConcatenate {
		return NewStringValue(left.String() + right.String()), nil
	}

	if op&OperatorTypeMask == OperatorTypeCompare {
		switch op {
		case OperatorNotEqual:
			return NewBoolValue(!left.Equal(right)), nil
		case OperatorEqual:
			return NewBoolValue(left.Equal(right)), nil
		case OperatorGreater:
			if left.IsBool() || right.IsBool() {
				return nil, ErrWrongType
			}
			return NewBoolValue(left.Greater(right)), nil
		case OperatorGreaterOrEqual:
			if left.IsBool() || right.IsBool() {
				return nil, ErrWrongType
			}
			return NewBoolValue(left.Greater(right) || left.Equal(right)), nil
		case OperatorLess:
			if left.IsBool() || right.IsBool() {
				return nil, ErrWrongType
			}
			return NewBoolValue(left.Less(right)), nil
		case OperatorLessOrEqual:
			if left.IsBool() || right.IsBool() {
				return nil, ErrWrongType
			}
			return NewBoolValue(left.Less(right) || left.Equal(right)), nil
		}
	}

	if op&OperatorTypeMask == OperatorTypeUnary {
		switch op {
		case OperatorNot:
			if !right.IsBool() {
				return nil, ErrWrongType
			}
			return NewBoolValue(!right.Bool()), nil
		}

		return nil, ErrWrongOperator
	}

	// Type Addition or Multiply

	if !((left == nil || left.IsNumber()) && right.IsNumber()) {
		return nil, ErrWrongType
	}

	switch op {
	case OperatorPlus:
		if left == nil {
			return NewNumberValue(right.Number()), nil
		}
		return NewNumberValue(left.Number() + right.Number()), nil
	case OperatorMinus:
		if left == nil {
			return NewNumberValue(-right.Number()), nil
		}
		return NewNumberValue(left.Number() - right.Number()), nil
	case OperatorMultiply:
		if left == nil {
			return nil, errors.New("left expression is nil")
		}
		return NewNumberValue(left.Number() * right.Number()), nil
	case OperatorDivide:
		if right.Number() == 0 {
			return nil, errors.New("division by zero")
		}

		if left == nil {
			return nil, errors.New("left expression is nil")
		}
		return NewNumberValue(left.Number() / right.Number()), nil
	default:
		return nil, errors.New("wrong Operator")
	}
}

func (m *mathNode) Add(item interface{}) *mathNode {
	switch i := item.(type) {
	case Valuer:
		if m.Item == nil {
			m.Item = item
			break
		}

		if _, ok := m.Item.(Valuer); ok {
			m.Item = item
			break
		}

		if m.Right != nil {
			m.Right = m.Right.Add(item)
		} else {
			m.Right = new(mathNode)
			m.Right.Item = item
		}
	case Operator:
		if m.Item == nil {
			m.Item = item
			break
		}

		if o, ok := m.Item.(Operator); ok {
			if i.LessOrEqualThan(o) && o != OperatorNot {
				node := new(mathNode)
				node.Item = i
				node.Left = m
				return node
			} else {
				if m.Right == nil {
					m.Right = new(mathNode)
				}
				m.Right = m.Right.Add(item)
			}
		} else {
			m.Left = new(mathNode)
			m.Left.Item = m.Item
			m.Item = item
		}
	}

	return m
}
