package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
	linters "github.com/springernature/halfpipe/linters/actions"
	"github.com/springernature/halfpipe/renderers/actions"
)

func init() {
	rootCmd.AddCommand(actionsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Generates a GitHub Actions workflow",
	Run: func(cmd *cobra.Command, args []string) {
		renderer := actions.NewActions()
		man, controller := getManifestAndController(renderer)
		response := controller.Process(man)

		actionsLintResult := linters.ActionsLinter{}.Lint(man)

		outputErrorsAndWarnings(nil, append(response.LintResults, actionsLintResult))
		fmt.Println(response.ConfigYaml)
	},
}
