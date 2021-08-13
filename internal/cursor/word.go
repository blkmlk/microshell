package cursor

type word struct {
	text string

	next *word
	prev *word
}

func (w *word) Text() string {
	return w.text
}

func (w *word) SetText(text string) {
	w.text = text
}

func (w *word) IsSpace() bool {
	return len(w.text) > 0 && w.text[0] == ' '
}

func (w *word) Len() int {
	return len(w.text)
}

func (w *word) End() int {
	if w.Len() > 0 {
		return w.Len() - 1
	}

	return 0
}

func newWord() *word {
	return &word{
		text: "",
		next: nil,
		prev: nil,
	}
}
