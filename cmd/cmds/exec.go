package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shell"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec <task name>",
	Short: "Prints command to execute the task locally",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			man, _ := getManifestAndController(formatInput(Input), nil)
			printAvailableExecTasks(man)
			os.Exit(0)
		}

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

func printAvailableExecTasks(man manifest.Manifest) {
	tasks := []string{}
	for _, t := range man.Tasks.Flatten() {
		switch t := t.(type) {
		case manifest.Run, manifest.DockerCompose, manifest.Buildpack:
			tasks = append(tasks, fmt.Sprintf("  %s", t.GetName()))
		}
	}
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No executable tasks found. Supported task types are run, docker-compose and buildpack.")
		return
	}
	fmt.Fprintf(os.Stderr, "Usage: halfpipe exec <task name>\n\nAvailable tasks:\n%s\n", strings.Join(tasks, "\n"))
}
