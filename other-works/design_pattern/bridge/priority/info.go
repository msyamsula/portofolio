package priority

import "fmt"

type Info struct{}

func (c Info) Format(m string) string {
	return fmt.Sprintf("Info, %s", m)
}
