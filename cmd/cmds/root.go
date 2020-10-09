package cmds

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/concourse"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "halfpipe",
	Short: `halfpipe is a tool to lint and render pipelines
Invoke without any arguments to lint your .halfpipe.io file and render a Concourse pipeline`,
	Run: func(cmd *cobra.Command, args []string) {
		render(concourse.NewPipeline(cfManifest.ReadAndInterpolateManifest))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErr(err)
		os.Exit(1)
	}
}
