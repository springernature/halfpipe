package cmds

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/renderers/actions"
	"github.com/springernature/halfpipe/renderers/concourse"
	"os"
)

func init() {
	rootCmd.AddCommand(urlCmd)
}

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Prints the pipeline url",
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.Afero{Fs: afero.NewOsFs()}

		currentDir, err := os.Getwd()
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false, formatInput(Input))
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		man, _ := getManifest(fs, currentDir, projectData.HalfpipeFilePath)

		if man.Platform.IsConcourse() {
			fmt.Println(concourse.NewPipeline(projectData.HalfpipeFilePath).PlatformURL(man))
		} else {
			fmt.Println(actions.NewActions(projectData.GitURI, projectData.HalfpipeFilePath).PlatformURL(man))
		}
	},
}
