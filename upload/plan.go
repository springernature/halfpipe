package upload

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Plan []Command

func (p Plan) Execute(stdout io.Writer, stderr io.Writer, stdin io.Reader, nonInteractive bool) (err error) {
	fmt.Fprintln(stdout, "Planned execution") // #nosec
	for _, cmd := range p {
		fmt.Fprintf(stdout, "\t* %s\n", cmd) // #nosec
	}

	if !nonInteractive {
		fmt.Fprint(stdout, "\nAre you sure? [y/N]: ") // #nosec
		var input string
		fmt.Fscan(stdin, &input) // #nosec
		if input != "y" {
			err = errors.New("aborted")
			return
		}
	}

	for _, cmd := range p {
		fmt.Fprintf(stdout, "\n=== %s ===\n", cmd) // #nosec
		if runErr := cmd.Run(stdout, stderr, stdin); runErr != nil {
			return runErr
		}
	}

	return
}

type Command struct {
	Cmd       exec.Cmd
	Printable string

	Executor func(stdout io.Writer, stdin io.Reader) error
}

func (c Command) String() string {
	if c.Printable != "" {
		return c.Printable
	}
	return strings.Join(c.Cmd.Args, " ")
}

func (c Command) Run(stdout io.Writer, stderr io.Writer, stdin io.Reader) error {
	if c.Executor != nil {
		if err := c.Executor(stdout, stdin); err != nil {
			return err
		}
	} else {
		if c.Cmd.Stdin == nil {
			c.Cmd.Stdin = stdin
		}
		if c.Cmd.Stdout == nil {
			c.Cmd.Stdout = stdout
		}
		if c.Cmd.Stderr == nil {
			c.Cmd.Stderr = stderr
		}
		if err := c.Cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
