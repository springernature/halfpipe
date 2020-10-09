package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render concourse pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndCreateController()

		pipelineConfig, lintResults := controller.Process(man)
		printErrAndResultAndExitOnError(nil, lintResults)

		fmt.Println(pipelineConfig)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
