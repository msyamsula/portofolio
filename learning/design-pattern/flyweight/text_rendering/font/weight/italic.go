package weight

var italicSingleton *Italic = nil

type Italic struct{}

func NewItalic() *Italic {
	if italicSingleton == nil {
		italicSingleton = &Italic{}
	}
	return italicSingleton
}

func (*Italic) IsFontWeight() bool {
	return true
}

func (*Italic) IsFontFlyweight() bool {
	return true
}
