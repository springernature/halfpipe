package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/blang/semver"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/controller"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/sync/githubRelease"
	"github.com/springernature/halfpipe/vault"
)

var (
	// This field will be populated in Concourse from the version resource
	// go build -ldflags "-X main.version`cat version/version`"
	version string
)

func main() {
	checkVersion()

	fs := afero.Afero{Fs: afero.NewOsFs()}
	ctrl := controller.Controller{
		Fs: fs,
		Linters: []linters.Linter{
			linters.TeamLinter{},
			linters.RepoLinter{},
			linters.SecretsLinter{vault.Vault{}, "concourse"},
			linters.TaskLinter{Fs: fs},
		},
		Renderer: pipeline.Pipeline{},
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
	currentVersion, err := getVersion()
	printAndExit(err)

	syncer := sync.Syncer{CurrentVersion: currentVersion, GithubRelease: githubRelease.GithubRelease{}}
	if len(os.Args) == 1 {
		printAndExit(syncer.Check())
	} else if len(os.Args) > 1 && os.Args[1] == "sync" {
		printAndExit(syncer.Update())
	}
}

func getVersion() (semver.Version, error) {
	if version == "" {
		return sync.DevVersion, nil
	}
	version, err := semver.Make(version)
	if err != nil {
		return semver.Version{}, err
	}
	return version, nil
}

func printAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		syscall.Exit(-1)
	}
}
