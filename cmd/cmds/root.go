package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/concourse"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a concourse pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndCreateController()

		pipelineConfig, lintResults := controller.Process(man)
		printErrAndResultAndExitOnError(nil, lintResults)

		pipeline, renderError := concourse.ToString(pipelineConfig)
		printErrAndResultAndExitOnError(renderError, nil)

		fmt.Println(pipeline)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
