package family

type FontFamily interface {
	IsFontFamily() bool
	IsFontFlyweight() bool
}

var familyRegistry = map[string]func() FontFamily{
	"arial":         func() FontFamily { return NewArial() },
	"timesnewroman": func() FontFamily { return NewTimesNewRoman() },
	"wingdings":     func() FontFamily { return NewWingdings() },
}

func NewFontFamily(fontFamily string) FontFamily {
	return familyRegistry[fontFamily]()
}
