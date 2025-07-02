package family

var timenewromansingleton *TimesNewRoman = nil

type TimesNewRoman struct{}

func NewTimesNewRoman() *TimesNewRoman {
	if timenewromansingleton == nil {
		timenewromansingleton = &TimesNewRoman{}
	}
	return timenewromansingleton
}

func (*TimesNewRoman) IsFontFamily() bool {
	return true
}

func (*TimesNewRoman) IsFontFlyweight() bool {
	return true
}
