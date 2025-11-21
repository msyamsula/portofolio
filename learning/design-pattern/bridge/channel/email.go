package channel

import "fmt"

type Email struct{}

func (e Email) Send(m string) string {
	return fmt.Sprintf("%s was send by email", m)
}
