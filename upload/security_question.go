package upload

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
)

var ErrWrongAnswer = errors.New("incorrect or empty response")

func SecurityQuestion(pipeline, currentBranch string) Command {
	pipelineName := fmt.Sprintf("%s-%s", pipeline, currentBranch)
	return Command{
		Printable: "# Security question",
		Executor: func(stdout io.Writer, stdin io.Reader) error {
			fmt.Fprintf(stdout, `WARNING! You are running halfpipe on a branch. WARNING!
* You are on branch %s
* We will upload the pipeline as %s
* Have you made sure any Cloud Foundry manifests you are using in deploy-cf tasks have different app name and routes than on the master branch? 
* Have you read the docs at https://docs.halfpipe.io/branches 
[y/N]: `, currentBranch, pipelineName)

			var input string
			fmt.Fscan(stdin, &input) // #nosec

			for _, expectedAnswer := range []string{"y", "yes", "Y", "Yes", "YES"} {
				if input == expectedAnswer {
					return nil
				}
			}
			return ErrWrongAnswer
		},
	}
}
