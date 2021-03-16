package cmds

import (
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/migrate"
)

func init() {
	rootCmd.AddCommand(actionsMigrationHelp)
}

var actionsMigrationHelp = &cobra.Command{
	Use:   "actions-migration-help",
	Short: "Prints out the steps needed to migrate from Concourse to Actions",
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndController(formatInput(Input))
		response := controller.Process(man)

		migrate.ActionsMigrationHelper(man, response)
	},
}
