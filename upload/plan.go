package upload

import (
	"errors"
	"fmt"
	"github.com/onsi/gomega/gbytes"
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
		fmt.Fscanln(stdin, &input) // #nosec
		if input != "y" {
			err = errors.New("aborted")
			return err
		}
	}

	for _, cmd := range p {
		fmt.Fprintf(stdout, "\n=== %s ===\n", cmd) // #nosec

		stderrOutput := gbytes.NewBuffer()
		capturingWriter := io.MultiWriter(stderr, stderrOutput)

		if runErr := cmd.Run(stdout, capturingWriter, stdin); runErr != nil {
			if cmd.ExecuteOnFailureFilter != nil && cmd.ExecuteOnFailureFilter(stderrOutput.Contents()) {
				for _, failureCmd := range cmd.ExecuteOnFailure {
					if runErr := failureCmd.Run(stdout, stderr, stdin); runErr != nil {
						return runErr
					}
				}
				continue
			}
			return runErr

		}
	}

	return err
}

type Command struct {
	Cmd       exec.Cmd
	Printable string

	Executor func(stdout io.Writer, stdin io.Reader) error

	ExecuteOnFailureFilter func(outputFromPreviousCommand []byte) bool
	ExecuteOnFailure       Plan
}

func (c Command) String() string {
	if c.Printable != "" {
		return c.Printable
	}

	s := strings.Join(c.Cmd.Args, " ")
	if len(c.ExecuteOnFailure) > 0 {
		s = fmt.Sprintf("%s || %s", s, c.ExecuteOnFailure[0].String())
	}
	return s
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
