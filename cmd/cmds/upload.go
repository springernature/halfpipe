package cmds

import (
	"os"
	"os/exec"
	"os/user"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/upload"
	"runtime"
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

			pipelineFile := func(fs afero.Afero) (afero.File, error) {
				return fs.OpenFile("pipeline.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			}

			currentBranch, err := project.BranchResolver()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			osResolver := func() string {
				return runtime.GOOS
			}

			currentDir, err := os.Getwd()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			planner := upload.NewPlanner(afero.Afero{Fs: afero.NewOsFs()}, exec.LookPath, currentUser.HomeDir, pipelineFile, nonInteractive, currentBranch, osResolver, os.Getenv, currentDir)

			plan, err := planner.Plan()
			if err != nil {
				printErr(err)
				os.Exit(1)
			}

			if err := plan.Execute(os.Stdout, os.Stderr, os.Stdin, nonInteractive); err != nil {
				printErr(err)
				os.Exit(1)
			}
		},
	}

	uploadCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "If this is set, you will not get prompted for action")
	rootCmd.AddCommand(uploadCmd)
}
