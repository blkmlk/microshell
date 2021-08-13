package terminal

type Output interface {
	Words(width, height int) []Word
}

type Word struct {
	text  string
	color Color
}

func NewWord(text string, color Color) Word {
	return Word{
		text:  text,
		color: color,
	}
}

func (o *Word) Len() int {
	return len(o.text)
}

func (o *Word) Text() string {
	return o.text
}

func (o *Word) SetText(text string) {
	o.text = text
}

func (o *Word) Color() Color {
	return o.color
}

func (o *Word) SetColor(color Color) {
	o.color = color
}
