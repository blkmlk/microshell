package terminal

const (
	maxFlexibleTableWidth = 150
	ftMinSpace            = 2
)

var _ Output = NewFlexibleTable()

func NewFlexibleTable() *flexibleTable {
	return &flexibleTable{}
}

type flexibleTable struct {
	words  []Word
	length int
}

func (t *flexibleTable) AddWord(w Word) {
	t.words = append(t.words, w)

	if t.length != 0 {
		t.length += ftMinSpace
	}

	t.length += w.Len()
}

func (t *flexibleTable) Words(width, height int) []Word {
	if len(t.words) == 0 {
		return nil
	}

	var result = make([]Word, 0, len(t.words)*2-1)

	var spaceWord Word
	space := ""
	for i := 0; i < ftMinSpace; i++ {
		space += " "
	}
	spaceWord.SetText(space)

	for i, w := range t.words {
		if i != 0 {
			result = append(result, spaceWord)
		}
		result = append(result, w)
	}

	return result
}
