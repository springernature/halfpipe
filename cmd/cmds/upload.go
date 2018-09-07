package cmds

import (
	"os"
	"os/exec"
	"os/user"

	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/upload"
	"io"
	"strings"
)

func init() {
	var nonInteractive bool

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Renders a pipeline and uploads it to halfpipe",
		Run: func(cmd *cobra.Command, args []string) {
			currentUser, err := user.Current()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			writer := &CapturingWriter{
				Stdout: os.Stdout,
			}

			pipelineFile := func(fs afero.Afero) (afero.File, error) {
				return fs.OpenFile("pipeline.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			}

			planner := upload.NewPlanner(afero.Afero{Fs: afero.NewOsFs()}, exec.LookPath, currentUser.HomeDir, writer, os.Stderr, os.Stdin, pipelineFile, nonInteractive)

			currentBranch, err := project.BranchResolver()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}
			if currentBranch != "master" {
				if secErr := planner.AskBranchSecurityQuestions(currentBranch); secErr != nil {
					printErr(secErr)
					os.Exit(1)
				}
			}

			plan, err := planner.Plan()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			if err := plan.Execute(writer, os.Stdin, nonInteractive); err != nil {
				printErr(err)
				os.Exit(1)
			}

			if shouldUnpause(writer) {
				fmt.Fprint(writer, "\n====================\n")                                                                            // #nosec
				fmt.Fprint(writer, "\nWhen the pipeline gets uploaded for the first time it must be unpaused. We will do it for you. \n") // #nosec

				plan, err := planner.Unpause()
				if err != nil {
					printErr(err)
					os.Exit(1)
				}

				if err := plan.Execute(writer, os.Stdin, true); err != nil {
					os.Exit(1)
				}
			}
		},
	}

	uploadCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "If this is set, you will not get prompted for action")
	rootCmd.AddCommand(uploadCmd)
}

func shouldUnpause(cw *CapturingWriter) bool {
	if strings.Contains(string(cw.BytesWritten), "the pipeline is currently paused. to unpause, either:") {
		return true
	}
	return false
}

type CapturingWriter struct {
	Stdout       io.Writer
	BytesWritten []byte
}

func (k *CapturingWriter) Write(p []byte) (n int, err error) {
	k.BytesWritten = append(k.BytesWritten, p...)
	return k.Stdout.Write(p)
}
