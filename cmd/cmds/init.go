package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates a sample halfpipe config",
	Run: func(cmd *cobra.Command, args []string) {
		currentDir, err := os.Getwd()
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		fs := afero.Afero{Fs: afero.NewOsFs()}
		projectResolver := project.NewProjectResolver(fs)

		err = manifest.NewSampleGenerator(fs, projectResolver, currentDir).Generate()

		if err != nil {
			printErr(err)
			os.Exit(1)
		}
		fmt.Println(fmt.Sprintf("Generated sample configuration at %s/.halfpipe.io", currentDir))

		return
	},
}
