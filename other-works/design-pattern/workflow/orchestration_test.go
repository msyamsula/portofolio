package workflow

import (
	"fmt"
	"testing"

	"github.com/msyamsula/portofolio/other-works/design-pattern/workflow/steps"
)

func TestOrchestration(t *testing.T) {
	// Create a new workflow orchestrator
	// next := func() steps.Step {

	// }
	// apiCall := steps.NewStep(next func ()steps.Step{}, procedure func(r int) error{})
	data := []int{}

	apiCall := steps.NewStep(
		"apiCall",
		steps.ProcedureFunc(func() error {
			data = []int{1, 2, 3, 4, 5}
			// fmt.Println("API call executed with data:", data)
			return nil
		}),
		3,
	)

	approval := steps.NewStep(
		"approve",
		steps.ProcedureFunc(func() error {
			// fmt.Println(data, "approved")
			if len(data) < 2 {
				// fmt.Println("goes error")
				// return errors.New("empty data")
				return fmt.Errorf("data is empty, cannot proceed with approval")

			}

			return nil
		}),
		1,
	)

	// reject := steps.NewStep(
	// 	"reject",
	// 	steps.ProcedureFunc(func() error {
	// 		// fmt.Println(data, "rejected")
	// 		return nil
	// 	}),
	// 	1,
	// )

	finish := steps.NewStep(
		"finish",
		steps.ProcedureFunc(func() error {
			// fmt.Println(data, "rejected")
			return nil
		}),
		1,
	)

	apiCall.SetNext(approval)
	approval.SetNext(finish)

	// fmt.Println(apiCall)

	root := apiCall

	err := root.Execute(0)
	fmt.Println(data, err)

	// apiCall := steps.NewApi()
}
