package family

var arialSingleton *Arial = nil

type Arial struct{}

func NewArial() *Arial {
	if arialSingleton == nil {
		arialSingleton = &Arial{}
	}
	return arialSingleton

}

func (*Arial) IsFontFamily() bool {
	return true
}

func (*Arial) IsFontFlyweight() bool {
	return true
}
