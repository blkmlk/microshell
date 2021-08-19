package history

import (
	"strconv"
	"testing"

	"github.com/sarulabs/di/v2"
	"github.com/stretchr/testify/suite"
)

func TestHistory(t *testing.T) {
	suite.Run(t, new(historyTestSuite))
}

type historyTestSuite struct {
	suite.Suite
	history *history
}

func (t *historyTestSuite) SetupTest() {
	builder, err := di.NewBuilder()
	t.Require().NoError(err)

	err = builder.Add(Definition)
	t.Require().NoError(err)

	ctn := builder.Build()
	t.history = ctn.Get(DefinitionName).(*history)
}

func (t *historyTestSuite) TestHistoryLoad() {
	const maxValues = 10
	var values []string

	for i := 0; i < maxValues; i++ {
		values = append(values, strconv.Itoa(i))
	}

	t.history.Load(values)
	t.Require().Equal(maxValues+1, t.history.list.Len())

	i := maxValues - 1
	for t.history.Prev() {
		t.Require().Equal(strconv.Itoa(i), t.history.Value())
		i--
	}
	t.Require().Equal(-1, i)
}

func (t *historyTestSuite) TestHistoryPush() {
	t.history.Cursor().Flush()
	t.history.Cursor().WriteRune('1')
	t.Require().True(t.history.Push())
	t.Require().Equal(2, t.history.list.Len())

	t.Require().True(t.history.Prev())
	t.history.Cursor().Flush()
	t.history.Cursor().WriteRune('2')
	t.Require().Equal(2, t.history.list.Len())
	t.Require().Equal(" 2", t.history.Cursor().String())

	t.Require().True(t.history.Next())
	t.history.Cursor().Flush()
	t.history.Cursor().WriteRune('3')
	t.Require().True(t.history.Push())
	t.Require().Equal(3, t.history.list.Len())

	t.Require().Equal(" ", t.history.Cursor().String())
	t.Require().True(t.history.Prev())
	t.Require().Equal(" 3", t.history.Cursor().String())
	t.Require().True(t.history.Prev())
	t.Require().Equal(" 1", t.history.Value())
	t.Require().False(t.history.Prev())
}
