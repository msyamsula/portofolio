package weight

type FontWeight interface {
	IsFontWeight() bool
	IsFontFlyweight() bool
}

var weightRegistry = map[string]func() FontWeight{
	"bold":   func() FontWeight { return NewBold() },
	"italic": func() FontWeight { return NewItalic() },
}

func NewFontWeight(fontWeight string) FontWeight {
	return weightRegistry[fontWeight]()
}
