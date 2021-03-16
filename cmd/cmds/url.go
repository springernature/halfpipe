package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/project"
)

func init() {
	urlCmd.Flags().StringVarP(&input, "input", "i", "", "Sets the halfpipe filename to be used")
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

		halfpipeFilenameOptions := config.HalfpipeFilenameOptions
		if input != "" {
			if strings.Contains(input, string(os.PathSeparator)) {
				fmt.Printf("Input file '%s' must be in current directory\n", input)
				os.Exit(1)
			}
			halfpipeFilenameOptions = []string{input}
		}

		projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false, halfpipeFilenameOptions)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		man, _ := getManifest(fs, currentDir, projectData.HalfpipeFilePath)
		if man.Platform.IsConcourse() {
			fmt.Printf("%s/teams/%s/pipelines/%s\n", config.ConcourseURL, man.Team, man.PipelineName())
		} else {
			url := strings.Replace(projectData.GitURI, "git@github.com:", "https://github.com/", 1)
			url = strings.TrimSuffix(url, ".git")
			fmt.Printf("%s/actions?query=workflow:%s\n", url, man.PipelineName())
		}

	},
}
