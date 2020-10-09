package cmds

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/concourse"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render concourse pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a pipeline`,
	Run: func(cmd *cobra.Command, args []string) {

		concourseRenderer := concourse.NewPipeline(cfManifest.ReadAndInterpolateManifest)

		man, controller := getManifestAndCreateController(concourseRenderer)

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
