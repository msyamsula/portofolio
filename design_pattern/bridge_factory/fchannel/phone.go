package fchannel

import "fmt"

type Phone struct{}

func (e Phone) Send(m string) string {
	return fmt.Sprintf("%s was send by phone", m)
}

func NewPhone() Phone {
	return Phone{}
}
