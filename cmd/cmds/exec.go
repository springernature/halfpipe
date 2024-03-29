package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/shell"
	"os"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec <task name>",
	Short: "Execute a task locally",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		shellRenderer := shell.New(taskName)
		man, controller := getManifestAndController(formatInput(Input), shellRenderer)

		response, err := controller.Process(man)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}
		outputLintResults(response.LintResults)
		fmt.Println(response)
	},
}
