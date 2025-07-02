package size

var mediumsingleton *Medium = nil

type Medium struct{}

func NewMedium() *Medium {
	if mediumsingleton == nil {
		mediumsingleton = &Medium{}
	}
	return mediumsingleton
}

func (*Medium) IsFontSize() bool {
	return true
}

func (*Medium) IsFontFlyweight() bool {
	return true
}
