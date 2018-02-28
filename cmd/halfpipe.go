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
	"github.com/springernature/halfpipe/sync/githubRelease"
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

func main() {
	if invokedForHelp(os.Args) {
		printHelpAndExit()
	}

	checkVersion()

	fs := afero.Afero{Fs: afero.NewOsFs()}

	currentDir, err := os.Getwd()
	printAndExit(err)

	proj, err := project.NewConfig(fs).Parse(currentDir)
	printAndExit(err)

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
			fmt.Println(err)
		}
		syscall.Exit(1)
	}

	pipelineYaml, err := pipeline.ToString(pipelineConfig)
	printAndExit(err)

	fmt.Println(pipelineYaml)
}

func checkVersion() {
	currentVersion, err := helpers.GetVersion()
	printAndExit(err)

	syncer := sync.Syncer{CurrentVersion: currentVersion, GithubRelease: githubRelease.GithubRelease{}}
	if len(os.Args) == 1 {
		printAndExit(syncer.Check())
	} else if len(os.Args) > 1 && os.Args[1] == "sync" {
		printAndExit(syncer.Update())
	}
}

func printAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		syscall.Exit(-1)
	}
}
