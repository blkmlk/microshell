package parser

import (
	"testing"

	"github.com/sarulabs/di/v2"

	"github.com/blkmlk/microshell/internal/logger"
	"github.com/blkmlk/microshell/internal/models"
	"github.com/stretchr/testify/suite"
)

type MathTreeExpressionTestSuite struct {
	suite.Suite
	ctx SystemContext
}

func TestMathTree(t *testing.T) {
	suite.Run(t, new(MathTreeExpressionTestSuite))
}

func (t *MathTreeExpressionTestSuite) SetupTest() {
	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(
		Definition,
		DefinitionContext,
		DefinitionScope,
		DefinitionCommandTree,
		logger.Definition,
	)
	t.Require().NoError(err)

	ctn := builder.Build()

	t.ctx = ctn.Get(DefinitionNameRootScope).(SystemContext)
}

func (t *MathTreeExpressionTestSuite) TestMathTree_Add() {
	tree := t.buildTree([]interface{}{"2", OperatorMultiply, "2", OperatorPlus, "2", OperatorMultiply, "3"})

	ctx := t.ctx

	v, err := tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.IsNumber())
	t.Require().Equal(10, v.Number())

	tree = t.buildTree([]interface{}{"2", OperatorConcatenate, "3"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().Equal("23", v.String())

	tree = t.buildTree([]interface{}{"2", OperatorMultiply, "3", OperatorConcatenate, "3"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().Equal("63", v.String())

	tree = t.buildTree([]interface{}{"1", OperatorPlus, "3", OperatorConcatenate, "10"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().Equal("410", v.String())

	tree = t.buildTree([]interface{}{"1", OperatorPlus, "3", OperatorEqual, "4"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{"10", OperatorMultiply, "3", OperatorNotEqual, "30"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().False(v.Bool())

	tree = t.buildTree([]interface{}{"10", OperatorMultiply, "3", OperatorEqual, "15", OperatorPlus, "20", OperatorMinus, "5"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{"10", OperatorMultiply, "3", OperatorEqual, "15", OperatorPlus, "20", OperatorMinus, "5"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{"true", OperatorGreater, "false"})
	_, err = tree.Value(ctx)
	t.Require().Error(err)

	tree = t.buildTree([]interface{}{"true", OperatorNotEqual, "false"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{OperatorNot, "false"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{"true", OperatorEqual, OperatorNot, "false"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{OperatorNot, "true", OperatorEqual, "false"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{"1", OperatorEqual, "1", OperatorEqual, "true"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())

	tree = t.buildTree([]interface{}{OperatorNot, OperatorNot, "true"})
	v, err = tree.Value(ctx)
	t.Require().NoError(err)
	t.Require().True(v.Bool())
}

func (t *MathTreeExpressionTestSuite) buildTree(items []interface{}) *mathTree {
	tree := NewMathTree()

	ctx := t.ctx

	for _, i := range items {
		switch t := i.(type) {
		case string:
			exp := NewStdExpression(true)

			for _, c := range t {
				exp.Add(ctx, models.Rune(c))
			}

			tree.Add(exp)
		case Operator:
			tree.Add(t)
		}
	}

	return tree
}
