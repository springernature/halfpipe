package cmds

import (
	"github.com/spf13/cobra"
	"os"
	"path"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndController(formatInput(Input), nil)
		response, err := controller.Process(man)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		if man.Platform.IsActions() && output == "" {
			output = path.Join(response.Project.GitRootPath, ".github/workflows/", man.PipelineName()+".yml")
		}

		renderResponse(response, output)
	},
}

var output string

func Execute() {
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Sets the path where the rendered pipeline will be saved to")
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
