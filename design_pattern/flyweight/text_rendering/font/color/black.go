package color

var blackSingleton *Black = nil

type Black struct{}

func NewBlack() *Black {
	if blackSingleton == nil {
		blackSingleton = &Black{}
	}

	return blackSingleton
}

func (*Black) IsFontColor() bool {
	return true
}

func (*Black) IsFontFlyweight() bool {
	return true
}
