package fchannel

import "fmt"

type Email struct{}

func (e Email) Send(m string) string {
	return fmt.Sprintf("%s was send by email", m)
}

func NewEmail() Email {
	return Email{}
}
