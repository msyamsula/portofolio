package color

type FontColor interface {
	IsFontColor() bool
	IsFontFlyweight() bool
}

var colorRegistry = map[string]func() FontColor{
	"black": func() FontColor {
		return NewBlack()
	},
	"blue": func() FontColor {
		return NewBlue()
	},
	"red": func() FontColor {
		return NewRed()
	},
}

func NewColor(colorType string) FontColor {
	return colorRegistry[colorType]()
}
