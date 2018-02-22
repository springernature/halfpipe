package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/controller"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/helpers/path_to_git"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/sync/githubRelease"
	"github.com/springernature/halfpipe/vault"
	"github.com/tcnksm/go-gitconfig"
)

var (
	// These field will be populated in Concourse
	// go build -ldflags "-X main.version=..."
	version     string
	vaultPrefix string
)

func main() {
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
			linters.SecretsLinter{VaultClient: vault.NewVaultClient(vaultPrefix)},
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
	gitUri, repoName, error := gitConfig()
	if error != nil {
		return
	}

	basePath, error := path_to_git.PathRelativeToGit(fs, currentDir, 5)
	if error != nil {
		return
	}

	project.GitUri = gitUri
	project.RepoName = repoName
	project.BasePath = basePath
	return
}

func gitConfig() (origin string, repoName string, error error) {
	origin, err := gitconfig.OriginURL()
	if err != nil {
		error = errors.New("Looks like you are not executing halfpipe from within a git repo?")
		return
	}

	repoName, _ = gitconfig.Repository()
	return
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
