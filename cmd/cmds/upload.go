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

			capturingStdOut := &CapturingWriter{
				Stdout: os.Stdout,
			}

			pipelineFile := func(fs afero.Afero) (afero.File, error) {
				return fs.OpenFile("pipeline.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			}

			currentBranch, err := project.BranchResolver()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			planner := upload.NewPlanner(afero.Afero{Fs: afero.NewOsFs()}, exec.LookPath, currentUser.HomeDir, pipelineFile, nonInteractive, currentBranch)

			plan, err := planner.Plan()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			if err := plan.Execute(capturingStdOut, os.Stderr, os.Stdin, nonInteractive); err != nil {
				printErr(err)
				os.Exit(1)
			}

			if shouldUnpause(capturingStdOut) {
				fmt.Fprint(capturingStdOut, "\n====================\n")                                                                            // #nosec
				fmt.Fprint(capturingStdOut, "\nWhen the pipeline gets uploaded for the first time it must be unpaused. We will do it for you. \n") // #nosec

				plan, err := planner.Unpause()
				if err != nil {
					printErr(err)
					os.Exit(1)
				}

				if err := plan.Execute(capturingStdOut, os.Stderr, os.Stdin, true); err != nil {
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
