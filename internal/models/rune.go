package models

type Rune rune

func (r Rune) Is(rn rune) bool {
	return rune(r) == rn
}

func (r Rune) IsSpace() bool {
	return rune(r) == ' '
}

func (r Rune) IsNumber() bool {
	return rune(r) >= '0' && rune(r) <= '9'
}

func (r Rune) IsLowerAlpha() bool {
	return rune(r) >= 'a' && rune(r) <= 'z'
}

func (r Rune) IsUpperAlpha() bool {
	return rune(r) >= 'A' && rune(r) <= 'Z'
}

func (r Rune) IsAlpha() bool {
	return r.IsLowerAlpha() || r.IsUpperAlpha()
}

func (r Rune) String() string {
	return string(r)
}
