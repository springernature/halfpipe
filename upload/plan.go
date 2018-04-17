package upload

import (
	"fmt"
	"os/exec"
)

type Plan []Command

type Command struct {
	Cmd       *exec.Cmd
	Printable string
}

func (p Plan) PrintPlan() {
	fmt.Println("Going to execute")
	for _, p := range p {
		fmt.Println(fmt.Sprintf("\t* %s", p.Printable))
	}
}

func (p Plan) Execute() (err error) {
	for _, cmd := range p {
		fmt.Println(fmt.Sprintf("=== %s ===", cmd.Printable))

		if runErr := cmd.Cmd.Start(); runErr != nil {
			err = runErr
			return
		}

		if waitErr := cmd.Cmd.Wait(); waitErr != nil {
			err = waitErr
			return
		}
	}

	return
}
