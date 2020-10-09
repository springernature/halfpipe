package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/pipeline/actions"
	"github.com/springernature/halfpipe/pipeline/concourse"
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

		var pipeline string
		var renderErr error
		if man.FeatureToggles.GithubActions() {
			pipeline, renderErr = actions.ToString(pipelineConfig.ActionsConfig)
		} else {
			pipeline, renderErr = concourse.ToString(pipelineConfig.ConcourseConfig)
		}

		printErrAndResultAndExitOnError(renderErr, nil)
		fmt.Println(pipeline)

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
