package fpriority

import "fmt"

type Critical struct{}

func (c Critical) Format(m string) string {
	return fmt.Sprintf("Critical, %s", m)
}

func NewCritical() Critical {
	return Critical{}
}
