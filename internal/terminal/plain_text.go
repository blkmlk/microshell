package terminal

var _ Output = NewPlainText("")

func NewPlainText(text string) *plainText {
	return &plainText{text: text}
}

type plainText struct {
	text string
}

func (p *plainText) SetText(text string) {
	p.text = text
}

func (p *plainText) Words(width, height int) []Word {
	t := p.text
	l := len(p.text)
	result := make([]Word, 0, l/width+1)

	for l > width {
		result = append(result, NewWord(t, ColorWhite), NewWord("\n", ColorWhite))
		l -= width
		t = t[width:]
	}

	if t != "" {
		result = append(result, NewWord(t, ColorWhite))
	}
	return result
}
