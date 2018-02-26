package main

import (
	"fmt"
	"os"
	"syscall"

	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/controller"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/helpers/path_to_git"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/sync/githubRelease"
	"github.com/springernature/halfpipe/vault"
	"github.com/tcnksm/go-gitconfig"
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

	projectData, err := projectData(fs, currentDir)
	printAndExit(err)

	ctrl := controller.Controller{
		Fs:      fs,
		Project: projectData,
		Linters: []linters.Linter{
			linters.TeamLinter{},
			linters.RepoLinter{Fs: fs},
			linters.SecretsLinter{VaultClient: vault.NewVaultClient(config.VaultPrefix)},
			linters.TaskLinter{Fs: fs},
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

func projectData(fs afero.Afero, currentDir string) (project model.Project, error error) {
	_, err := exec.LookPath("git")

	if err != nil {
		error = errors.New("Looks like you don't have git installed? please make sure you do")
		return
	}

	origin, err := gitconfig.OriginURL()
	if err != nil {
		error = errors.New("Looks like you are not executing halfpipe from within a git repo?")
		return
	}

	basePath, error := path_to_git.PathRelativeToGit(fs, currentDir, 5)
	if error != nil {
		return
	}

	project.GitUri = origin
	project.BasePath = basePath
	return
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
