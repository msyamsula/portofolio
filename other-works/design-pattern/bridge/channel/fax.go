package channel

import "fmt"

type Fax struct{}

func (e Fax) Send(m string) string {
	return fmt.Sprintf("%s was send by fax", m)
}
