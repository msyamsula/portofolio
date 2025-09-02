package steps

import (
	"errors"
	"fmt"
)

type work struct {
	baseStep
}

func NewWork() *work {
	return &work{}
}

func (*work) IsStep() bool {
	return true
}

func (a *work) SetNext(next Step) {
	a.next = next
}

func (a *work) SetProcedure(p ProcedureFunc) {
	a.procedure = p
}

func (a *work) Execute(retry int) error {
	fmt.Println("Executing step:", a.name, "with retry count:", retry)
	if retry == a.maxRetry {
		a.state = "error"
		return errors.New("max retries reached") // or return an error indicating max retries reached
	}
	err := a.procedure()
	if err != nil {
		a.state = "retried"
		return a.Execute(retry + 1) // retry with incremented retry count
	}

	a.state = "success"
	fmt.Println("Step executed successfully:", a.name)
	if a.next == nil {
		return nil // no next step, return success
	}
	return a.next.Execute(0)

}
