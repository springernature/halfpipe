package cmds

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a Concourse pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndController(nil)
		response := controller.Process(man)

		var outputPath string
		if man.FeatureToggles.GithubAction() {
			outputPath = path.Join(response.Project.GitRootPath, ".github/workflows/", man.PipelineName()+".yml")
		}

		renderResponse(response, outputPath)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
