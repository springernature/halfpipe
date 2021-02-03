package cmds

import (
	"path"

	"github.com/spf13/cobra"
	linters "github.com/springernature/halfpipe/linters/actions"
	"github.com/springernature/halfpipe/renderers/actions"
)

func init() {
	rootCmd.AddCommand(actionsCmd)
	actionsCmd.Flags().StringVarP(&outputPath, "output", "o", "", "override the default output filepath")
}

var outputPath string

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Generates a GitHub Actions workflow",
	Run: func(cmd *cobra.Command, args []string) {
		renderer := actions.NewActions()
		man, controller := getManifestAndController(renderer)
		response := controller.Process(man)

		actionsLintResult := linters.ActionsLinter{}.Lint(man)
		response.LintResults = append(response.LintResults, actionsLintResult)

		if outputPath == "" {
			outputPath = path.Join(response.Project.GitRootPath, ".github/workflows/", man.PipelineName()+".yml")
		}
		renderResponse(response, outputPath)
	},
}
