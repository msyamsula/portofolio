package size

var smallsingleton *Small = nil

type Small struct{}

func NewSmall() *Small {
	if smallsingleton == nil {
		smallsingleton = &Small{}
	}
	return smallsingleton
}

func (*Small) IsFontSize() bool {
	return true
}

func (*Small) IsFontFlyweight() bool {
	return true
}
