package size

type FontSize interface {
	IsFontSize() bool
	IsFontFlyweight() bool
}

var sizeRegistry = map[string]func() FontSize{
	"small":  func() FontSize { return NewSmall() },
	"medium": func() FontSize { return NewMedium() },
	"large":  func() FontSize { return NewLarge() },
}

func NewFontSize(fontSize string) FontSize {
	return sizeRegistry[fontSize]()
}
