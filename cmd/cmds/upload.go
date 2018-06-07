package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/upload"
)

func init() {
	rootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Renders a pipeline and uploads it to halfpipe",
	Run: func(cmd *cobra.Command, args []string) {
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		pipelineFile := func(fs afero.Afero) (afero.File, error) {
			return fs.OpenFile("pipeline.yml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		}

		planner := upload.NewPlanner(afero.Afero{Fs: afero.NewOsFs()}, exec.LookPath, currentUser.HomeDir, os.Stdout, os.Stderr, os.Stdin, pipelineFile)

		plan, err := planner.Plan()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := plan.Execute(os.Stdout, os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	},
}
