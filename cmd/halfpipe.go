package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/blang/semver"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/controller"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/sync/githubRelease"
)

var (
	// This field will be populated in Concourse from the version resource
	// go build -ldflags "-X main.version`cat version/version`"
	version string
)

func getVersion() (semver.Version, error) {
	if version == "" {
		return halfpipe.DevVersion, nil
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

func main() {
	checkVersion()

	ctrl := controller.Controller{
		Fs: afero.Afero{Fs: afero.NewOsFs()},
		Linters: []linters.Linter{
			linters.TeamLinter{},
			linters.RepoLinter{},
		},
	}

	pipeline, errs := ctrl.Process()
	if len(errs) > 0 {
		println("there were some errors")
		for _, err := range errs {
			fmt.Println(err)
		}
	}

	fmt.Println(pipeline)
}

func checkVersion() {
	currentVersion, err := getVersion()
	printAndExit(err)

	syncer := sync.Syncer{CurrentVersion: currentVersion, GithubRelease: githubRelease.GithubRelease{}}
	if len(os.Args) == 1 {
		printAndExit(syncer.Check())
	} else if len(os.Args) > 1 && os.Args[1] == "sync" {
		printAndExit(syncer.Update())
		return
	}
}
