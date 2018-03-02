package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/controller"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/secret_resolver"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/project"
	"github.com/springernature/halfpipe/sync"
)

func invokedForHelp(args []string) bool {
	return len(args) > 1 && (args[1] == "-h" || args[1] == "-help" || args[1] == "--help")

}

func printHelpAndExit() {
	version, _ := helpers.GetVersion()
	fmt.Println("Sup! Docs are at https://docs.halfpipe.io")
	fmt.Printf("Current version is %s\n", version)
	fmt.Println("Available commands are")
	fmt.Printf("\tsync - updates the halfpipe cli to latest version `halfpipe sync`\n")
	syscall.Exit(0)
}

func invokedForSync(args []string) bool {
	return len(args) > 1 && args[1] == "sync"

}

func syncBinary() (err error) {
	currentVersion, err := helpers.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion)
	err = syncer.Update(os.Stdout)
	return
}

func lintAndRender() (err error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}

	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	proj, err := project.NewConfig(fs).Parse(currentDir)
	if err != nil {
		return
	}

	ctrl := controller.Controller{
		Fs:      fs,
		Project: proj,
		Linters: []linters.Linter{
			linters.TeamLinter{},
			linters.RepoLinter{Fs: fs},
			linters.SecretsLinter{
				ConcourseResolv: secret_resolver.NewConcourseResolver(config.VaultPrefix, secret_resolver.NewSecretResolver(fs)),
			},
			linters.TaskLinter{Fs: fs},
			linters.ArtifactsLinter{},
		},
		Renderer:  pipeline.Pipeline{},
		Defaulter: defaults.DefaultValues.Update,
	}

	pipelineConfig, lintResults := ctrl.Process()
	if lintResults.HasErrors() {
		for _, err := range lintResults {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	pipelineYaml, err := pipeline.ToString(pipelineConfig)
	if err != nil {
		return
	}

	fmt.Println(pipelineYaml)
	return
}

func main() {
	err := checkVersion()
	if err == nil {
		if invokedForHelp(os.Args) {
			printHelpAndExit()
		} else if invokedForSync(os.Args) {
			err = syncBinary()
		} else {
			lintAndRender()
		}
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		syscall.Exit(-1)
	}
}

func checkVersion() (err error) {
	currentVersion, err := helpers.GetVersion()
	if err != nil {
		return
	}

	syncer := sync.NewSyncer(currentVersion)
	err = syncer.Check()
	return
}
