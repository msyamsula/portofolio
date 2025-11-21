package channel

import "fmt"

type Phone struct{}

func (e Phone) Send(m string) string {
	return fmt.Sprintf("%s was send by phone", m)
}
