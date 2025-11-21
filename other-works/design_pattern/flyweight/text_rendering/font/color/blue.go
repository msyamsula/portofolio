package color

var blueSingleton *Blue = nil

type Blue struct{}

func NewBlue() *Blue {
	if blueSingleton == nil {
		blueSingleton = &Blue{}
	}
	return blueSingleton
}
func (*Blue) IsFontColor() bool {
	return true
}

func (*Blue) IsFontFlyweight() bool {
	return true
}
