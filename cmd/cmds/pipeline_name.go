package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pipelineNameCmd)
}

var pipelineNameCmd = &cobra.Command{
	Use:   "pipeline-name",
	Short: "Prints the name of the pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		man, _ := getManifestAndController()
		if man.PipelineName() == "" {
			os.Exit(1)
		}
		fmt.Println(man.PipelineName())
	},
}
