package upload

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Plan []Command

func (p Plan) Execute(stdout io.Writer, stdin io.Reader) (err error) {
	fmt.Fprintln(stdout, "Planned execution") // #nosec
	for _, cmd := range p {
		fmt.Fprintf(stdout, "\t* %s\n", cmd) // #nosec
	}
	fmt.Fprint(stdout, "\nAre you sure? [y/N]: ") // #nosec
	var input string
	fmt.Fscan(stdin, &input) // #nosec
	if input != "y" {
		err = errors.New("aborted")
	}

	for _, cmd := range p {
		fmt.Fprintf(stdout, "\n=== %s ===\n", cmd) // #nosec
		runErr := cmd.Cmd.Run()
		if runErr != nil {
			err = runErr
			return
		}
	}

	return
}

type Command struct {
	Cmd       exec.Cmd
	Printable string
}

func (c Command) String() string {
	if c.Printable != "" {
		return c.Printable
	}
	return strings.Join(c.Cmd.Args, " ")

}
