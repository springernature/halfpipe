package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/config"
	"os"
	"path"
	"strings"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		var halfpipeFilenameOptions []string
		if input == "" {
			halfpipeFilenameOptions = config.HalfpipeFilenameOptions
		} else {
			if strings.Contains(input, string(os.PathSeparator)) {
				fmt.Println(fmt.Sprintf("Input file '%s' must be in current directory", input))
				os.Exit(1)
			}
			halfpipeFilenameOptions = []string{input}
		}
		man, controller := getManifestAndController(halfpipeFilenameOptions)
		response := controller.Process(man)

		if man.Platform.IsActions() && output == "" {
			output = path.Join(response.Project.GitRootPath, ".github/workflows/", man.PipelineName()+".yml")
		}

		renderResponse(response, output)
	},
}

var input string
var output string

func Execute() {
	rootCmd.Flags().StringVarP(&input, "input", "i", "", "Sets the halfpipe filename to be used")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Sets the path where the rendered pipeline will be saved to")
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
