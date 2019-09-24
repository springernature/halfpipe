package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/project"
)

func init() {
	rootCmd.AddCommand(urlCmd)
}

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Prints the url of the Concourse pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.Afero{Fs: afero.NewOsFs()}

		currentDir, err := os.Getwd()
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		man, _ := getManifest(fs, currentDir, projectData.HalfpipeFilePath)
		fmt.Printf("%s/teams/%s/pipelines/%s\n", config.ConcourseURL, man.Team, man.PipelineName())
		return
	},
}
