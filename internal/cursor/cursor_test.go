package cursor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCursor_AllWords(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc edf", []actionTest{}, []string{"abc", " ", "edf"}, 0)
}

func TestCursor_Words(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc edf", []actionTest{}, []string{"abc", " ", "edf"}, 0)
	require.Equal(t, []string{"abc", "edf"}, w.Words())
}

func TestCursor_WordsFromCursor(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc edf", []actionTest{}, []string{"abc", " ", "edf"}, 0)

	w.MoveToStart()
	w.MoveToNextWord()
	w.MoveForward()
	require.Equal(t, []string{"edf"}, w.WordsFromCursor())
}

func TestCursor_AllWordsFromCursor(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc edf", []actionTest{}, []string{"abc", " ", "edf"}, 0)

	w.MoveToStart()
	w.MoveToNextWord()
	w.MoveForward()
	require.Equal(t, []string{" ", "edf"}, w.AllWordsFromCursor())
}

func TestCursor_StringFromPositionOffset(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc edf", []actionTest{}, []string{"abc", " ", "edf"}, 0)

	w.MoveToStart()
	w.MoveToNextWord()
	require.Equal(t, "c edf", w.StringFromPosition())
	require.Equal(t, " edf", w.StringFromPositionOffset(1))
	require.Equal(t, "bc edf", w.StringFromPositionOffset(-1))

	w.MoveForward()
	require.Equal(t, "c edf", w.StringFromPositionOffset(-1))

	w.MoveToStart()
	require.Equal(t, " abc edf", w.StringFromPositionOffset(-5))

	w.MoveToEnd()
	require.Equal(t, "f", w.StringFromPositionOffset(5))
}

func TestCursor_WriteRune(t *testing.T) {
	w := NewCursor()
	w.WriteRune('a')
	w.WriteRune('b')
	w.WriteRune('d')

	require.Equal(t, 1, w.MoveBackward())
	w.WriteRune('c')
	require.Equal(t, []string{"abcd"}, w.AllWords())

	require.Equal(t, 1, w.MoveBackward())
	w.WriteRune(' ')
	require.Equal(t, []string{"ab", " ", "cd"}, w.AllWords())

	w = NewCursor()
	testCase(t, w, "abc defg hijk", []actionTest{
		{w.MoveToNextWord, 3, "move_next_word"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
	}, []string{"abc", " ", "defg", " ", "hijk"}, 6)

	w.WriteRune(' ')
	require.Equal(t, []string{"abc", " ", "de", " ", "fg", " ", "hijk"}, w.AllWords())
}

func TestCursor_WriteString(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "abc", []actionTest{}, []string{"abc"}, 0)
}

func TestCursor_SetPosition(t *testing.T) {
	w := NewCursor().(*cursor)

	_, err := w.WriteString("123 456 789")
	require.NoError(t, err)

	w.SetPosition(6)
	require.Equal(t, 6, w.Position())
	require.Equal(t, 1, w.offset)
	require.Equal(t, "56 789", w.StringFromPosition())
}

func TestCursor_Swap(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.Swap, 0, "swap"},
	}, nil, 0)

	testCase(t, w, "a", []actionTest{
		{w.Swap, 0, "swap"},
	}, []string{"a"}, 0)

	testCase(t, w, "ab", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.Swap, 2, "swap"},
	}, []string{"ba"}, 2)

	testCase(t, w, "ab ", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Swap, 2, "swap"},
	}, []string{"a", " ", "b"}, 3)

	testCase(t, w, "abc", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Swap, 2, "swap"},
	}, []string{"acb"}, 3)
}

func TestCursor_MoveBackward(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveBackward, 0, "move_backward"},
	}, nil, 0)

	testCase(t, w, "  ", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveBackward, 1, "move_backward"},
	}, nil, 0)
}

func TestCursor_MoveForward(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveForward, 0, "move_forward"},
	}, nil, 0)
}

func TestCursor_MoveToStart(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveToStart, 0, "move_start"},
	}, nil, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToStart, 0, "move_start"},
	}, []string{"abc"}, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToEnd, 3, "move_end"},
		{w.MoveToStart, 3, "move_start"},
	}, []string{"abc"}, 0)

	testCase(t, w, "abc  ", []actionTest{
		{w.MoveToEnd, 5, "move_end"},
		{w.MoveToStart, 5, "move_start"},
	}, []string{"abc", "  "}, 0)

	testCase(t, w, "abc  ", []actionTest{
		{w.MoveToEnd, 5, "move_end"},
		{w.DeleteToPrevWord, 5, "delete_prev_word"},
		{w.MoveToStart, 0, "move_start"},
	}, nil, 0)
}

func TestCursor_MoveToEnd(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveToEnd, 0, "move_end"},
	}, nil, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToEnd, 3, "move_end"},
	}, []string{"abc"}, 3)

	testCase(t, w, "abc", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveToEnd, 2, "move_end"},
		{w.MoveToEnd, 0, "move_end"},
	}, []string{"abc"}, 3)
}

func TestCursor_MoveToPrevWord(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveToPrevWord, 0, "move_prev_word"},
	}, nil, 0)

	testCase(t, w, "  ", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveToPrevWord, 1, "move_prev_word"},
	}, nil, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToPrevWord, 0, "move_prev_word"},
	}, []string{"abc"}, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToEnd, 3, "move_to_end"},
		{w.MoveToPrevWord, 3, "move_prev_word"},
	}, []string{"abc"}, 0)

	testCase(t, w, "abc  ", []actionTest{
		{w.MoveToEnd, 5, "move_to_end"},
		{w.MoveToPrevWord, 5, "move_prev_word"},
	}, []string{"abc", "  "}, 0)

	testCase(t, w, "abc  efg", []actionTest{
		{w.MoveToEnd, 8, "move_to_end"},
		{w.MoveToPrevWord, 3, "move_prev_word"},
		{w.MoveToPrevWord, 5, "move_prev_word"},
		{w.MoveToPrevWord, 0, "move_prev_word"},
	}, []string{"abc", "  ", "efg"}, 0)

	testCase(t, w, "ab cd efg", []actionTest{
		{w.MoveToEnd, 9, "move_to_end"},
		{w.MoveToPrevWord, 3, "move_prev_word"},
	}, []string{"ab", " ", "cd", " ", "efg"}, 6)
}

func TestCursor_MoveToNextWord(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.MoveToNextWord, 0, "move_next_word"},
	}, nil, 0)

	testCase(t, w, "a b", []actionTest{
		{w.MoveToNextWord, 1, "move_next_word"},
	}, []string{"a", " ", "b"}, 1)

	testCase(t, w, "a  boo", []actionTest{
		{w.MoveToNextWord, 1, "move_next_word"},
		{w.MoveToNextWord, 5, "move_next_word"},
	}, []string{"a", "  ", "boo"}, 6)

	testCase(t, w, "aaa  boo", []actionTest{
		{w.MoveToNextWord, 3, "move_next_word"},
		{w.MoveToNextWord, 5, "move_next_word"},
	}, []string{"aaa", "  ", "boo"}, 8)

	testCase(t, w, "aaa  ", []actionTest{
		{w.MoveToNextWord, 3, "move_next_word"},
		{w.MoveToNextWord, 2, "move_next_word"},
	}, []string{"aaa", "  "}, 5)

	testCase(t, w, "    aaa ", []actionTest{
		{w.MoveToNextWord, 7, "move_next_word"},
		{w.MoveToNextWord, 1, "move_next_word"},
	}, []string{"aaa", " "}, 8)

	testCase(t, w, "aaa  ", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveToNextWord, 2, "move_next_word"},
		{w.MoveToNextWord, 2, "move_next_word"},
		{w.MoveToNextWord, 0, "move_next_word"},
	}, []string{"aaa", "  "}, 5)
}

func TestCursor_Backspace(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.Backspace, 0, "backspace"},
	}, nil, 0)

	testCase(t, w, "t", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.Backspace, 1, "backspace"},
	}, nil, 0)

	testCase(t, w, "a b", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Backspace, 1, "backspace"},
	}, []string{"ab"}, 1)

	testCase(t, w, "a b", []actionTest{
		{w.MoveToEnd, 3, "move_to_end"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 0, "backspace"},
	}, nil, 0)

	testCase(t, w, "123456789", []actionTest{
		{w.MoveToEnd, 9, "move_to_end"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
	}, []string{"1234"}, 4)

	testCase(t, w, "abcdefg", []actionTest{
		{w.MoveToEnd, 7, "move_to_end"},
		{w.MoveBackward, 1, "move_backward"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
		{w.Backspace, 1, "backspace"},
	}, []string{"abcg"}, 3)
}

func TestCursor_Delete(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.Delete, 0, "delete"},
	}, nil, 0)

	testCase(t, w, "a", []actionTest{
		{w.Delete, 1, "delete"},
	}, nil, 0)

	testCase(t, w, "ab", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.Backspace, 1, "backspace"},
		{w.Delete, 1, "delete"},
	}, nil, 0)

	testCase(t, w, "abc", []actionTest{
		{w.Delete, 1, "delete"},
	}, []string{"bc"}, 0)

	testCase(t, w, "abc", []actionTest{
		{w.MoveToEnd, 3, "move_forward"},
		{w.Delete, 0, "delete"},
	}, []string{"abc"}, 3)

	testCase(t, w, "abc edf", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Delete, 1, "delete"},
	}, []string{"abcedf"}, 3)

	testCase(t, w, "abc  edf", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Delete, 1, "delete"},
	}, []string{"abc", " ", "edf"}, 3)

	testCase(t, w, "abc", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.Delete, 1, "delete"},
	}, []string{"ac"}, 1)

	testCase(t, w, "abc", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.Delete, 1, "delete"},
	}, []string{"ab"}, 2)
}

func TestCursor_DeleteToPrevWord(t *testing.T) {
	w := NewCursor().(*cursor)

	testCase(t, w, "", []actionTest{
		{w.DeleteToPrevWord, 0, "delete_prev_word"},
	}, nil, 0)

	testCase(t, w, "   ", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToPrevWord, 1, "delete_prev_word"},
	}, nil, 0)
	require.Equal(t, 2, w.current.Len())

	testCase(t, w, "abc def   ", []actionTest{
		{w.MoveToEnd, 10, "move_to_end"},
		{w.MoveBackward, 1, "move_backward"},
		{w.DeleteToPrevWord, 5, "delete_prev_word"},
	}, []string{"abc", "  "}, 4)

	testCase(t, w, "abc def   gh", []actionTest{
		{w.MoveToEnd, 12, "move_to_end"},
		{w.MoveBackward, 1, "move_backward"},
		{w.MoveBackward, 1, "move_backward"},
		{w.DeleteToPrevWord, 6, "delete_prev_word"},
	}, []string{"abc", " ", "gh"}, 4)

	testCase(t, w, "abc", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToPrevWord, 1, "delete_prev_word"},
	}, []string{"bc"}, 0)

	testCase(t, w, "abc efg", []actionTest{
		{w.MoveToEnd, 7, "move_to_end"},
		{w.DeleteToPrevWord, 3, "delete_prev_word"},
	}, []string{"abc", " "}, 4)

	testCase(t, w, "abc efg", []actionTest{
		{w.MoveToEnd, 7, "move_forward"},
		{w.DeleteToPrevWord, 3, "delete_prev_word"},
		{w.DeleteToPrevWord, 4, "delete_prev_word"},
		{w.DeleteToPrevWord, 0, "delete_prev_word"},
	}, nil, 0)

	testCase(t, w, "a b  c", []actionTest{
		{w.MoveToNextWord, 1, "move_next_word"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToPrevWord, 1, "delete_prev_word"},
	}, []string{"a", "   ", "c"}, 2)

	testCase(t, w, "abc  ", []actionTest{
		{w.MoveToEnd, 5, "move_to_end"},
		{w.DeleteToPrevWord, 5, "delete_prev_word"},
	}, nil, 0)

	testCase(t, w, "t1             t2               t3                       ", []actionTest{
		{w.MoveToEnd, 57, "move_to_end"},
		{w.DeleteToPrevWord, 25, "delete_prev_word"},
		{w.DeleteToPrevWord, 17, "delete_prev_word"},
	}, []string{"t1", "             "}, 15)
}

func TestCursor_DeleteToStart(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.DeleteToStart, 0, "delete_to_start"},
	}, nil, 0)

	testCase(t, w, "t1", []actionTest{
		{w.DeleteToStart, 0, "delete_to_start"},
	}, []string{"t1"}, 0)

	testCase(t, w, "t1", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToStart, 1, "delete_to_start"},
	}, []string{"1"}, 0)

	testCase(t, w, "t1   t2", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToStart, 3, "delete_to_start"},
	}, []string{"  ", "t2"}, 0)

	testCase(t, w, "t1   t2", []actionTest{
		{w.MoveToEnd, 7, "move_to_end"},
		{w.DeleteToStart, 7, "delete_to_start"},
	}, nil, 0)
}

func TestCursor_DeleteToEnd(t *testing.T) {
	w := NewCursor()

	testCase(t, w, "", []actionTest{
		{w.DeleteToEnd, 0, "delete_to_end"},
	}, nil, 0)

	testCase(t, w, "t1 t2", []actionTest{
		{w.DeleteToEnd, 5, "delete_to_end"},
	}, nil, 0)

	testCase(t, w, "t1 t2", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToEnd, 4, "delete_to_end"},
	}, []string{"t"}, 1)

	testCase(t, w, "t1  t2", []actionTest{
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.MoveForward, 1, "move_forward"},
		{w.DeleteToEnd, 3, "delete_to_end"},
	}, []string{"t1", " "}, 3)

	testCase(t, w, "t1  t2", []actionTest{
		{w.MoveToEnd, 6, "move_to_end"},
		{w.DeleteToEnd, 0, "delete_to_end"},
	}, []string{"t1", "  ", "t2"}, 6)

	testCase(t, w, "t1  t2", []actionTest{
		{w.MoveToEnd, 6, "move_to_end"},
		{w.MoveBackward, 1, "move_backward"},
		{w.DeleteToEnd, 1, "delete_to_end"},
	}, []string{"t1", "  ", "t"}, 5)
}

type actionTest struct {
	action action
	result int
	name   string
}

func testCase(t *testing.T, c Cursor, initValue string, actionTests []actionTest, result []string, position int) {
	c.Flush()

	_, err := c.WriteString(initValue)
	require.NoError(t, err)
	c.MoveToStart()

	require.Equal(t, 0, c.Position(), "init position")

	for _, at := range actionTests {
		require.Equal(t, at.result, at.action(), at.name)
	}

	require.Equal(t, result, c.AllWords(), "all words")
	require.Equal(t, position, c.Position(), "position")
}
