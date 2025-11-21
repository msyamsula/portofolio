package steps

type ProcedureFunc func() error

// type nextFunc func() Step

type Step interface {
	IsStep() bool
	// Next() Step
	Execute(r int) error
	SetProcedure(ProcedureFunc)
	SetNext(Step)
}

type baseStep struct {
	name      string
	state     string
	next      Step
	procedure ProcedureFunc
	maxRetry  int
}

func NewStep(name string, procedure ProcedureFunc, maxRetries int) Step {
	w := &work{
		baseStep: baseStep{
			name:      name,
			state:     "",
			next:      nil,
			procedure: procedure,
			maxRetry:  maxRetries,
		},
	}
	// w.SetProcedure(procedure)
	return w
}
