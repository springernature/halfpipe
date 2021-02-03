package cmds

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/concourse"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a Concourse pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		renderer := concourse.NewPipeline()
		man, controller := getManifestAndController(renderer)
		response := controller.Process(man)
		renderResponse(response, "")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
