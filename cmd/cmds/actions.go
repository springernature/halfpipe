package cmds

import (
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/renderers/actions"
)

func init() {
	rootCmd.AddCommand(actionsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Generates a GitHub Actions workflow",
	Run: func(cmd *cobra.Command, args []string) {
		render(actions.NewActions())
	},
}
