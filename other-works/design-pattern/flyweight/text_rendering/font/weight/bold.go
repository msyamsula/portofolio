package weight

var boldSingleton *Bold = nil

type Bold struct{}

func NewBold() *Bold {
	if boldSingleton == nil {
		boldSingleton = &Bold{}
	}
	return boldSingleton
}

func (*Bold) IsFontWeight() bool {
	return true
}

func (*Bold) IsFontFlyweight() bool {
	return true
}
