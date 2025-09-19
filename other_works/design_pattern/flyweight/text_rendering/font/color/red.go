package color

type Red struct{}

var redSingleton *Red = nil

func NewRed() *Red {
	if redSingleton == nil {
		redSingleton = &Red{}
	}
	return redSingleton
}

func (*Red) IsFontColor() bool {
	return true
}

func (*Red) IsFontFlyweight() bool {
	return true
}
