package cmds

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	pipelineNameCmd.Flags().StringVarP(&input, "input", "i", "", "Sets the halfpipe filename to be used")
	rootCmd.AddCommand(pipelineNameCmd)
}

var pipelineNameCmd = &cobra.Command{
	Use:   "pipeline-name",
	Short: "Prints the name of the pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		halfpipeFilenameOptions := config.HalfpipeFilenameOptions
		if input != "" {
			if strings.Contains(input, string(os.PathSeparator)) {
				fmt.Printf("Input file '%s' must be in current directory\n", input)
				os.Exit(1)
			}
			halfpipeFilenameOptions = []string{input}
		}

		man, _ := getManifestAndController(halfpipeFilenameOptions)
		if man.PipelineName() == "" {
			os.Exit(1)
		}
		fmt.Println(man.PipelineName())
	},
}
