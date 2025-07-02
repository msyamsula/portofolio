package size

var largesingleton *Large = nil

type Large struct{}

func NewLarge() *Large {
	if largesingleton == nil {
		largesingleton = &Large{}
	}

	return largesingleton
}

func (*Large) IsFontSize() bool {
	return true
}

func (*Large) IsFontFlyweight() bool {
	return true
}
