package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/shell"
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

		shellRenderer := shell.NewShell(taskName)
		man, controller := getManifestAndController(formatInput(Input), shellRenderer)

		response := controller.Process(man)
		outputLintResults(response.LintResults)
		fmt.Println(response)

	},
}
