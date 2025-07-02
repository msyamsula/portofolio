package family

var wingdingssingleton *Wingdings = nil

type Wingdings struct{}

func NewWingdings() *Wingdings {
	if wingdingssingleton == nil {
		wingdingssingleton = &Wingdings{}
	}
	return wingdingssingleton
}

func (*Wingdings) IsFontFamily() bool {
	return true
}

func (*Wingdings) IsFontFlyweight() bool {
	return true
}
