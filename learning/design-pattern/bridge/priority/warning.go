package priority

import "fmt"

type Warning struct{}

func (c Warning) Format(m string) string {
	return fmt.Sprintf("Warning, %s", m)
}
