package main

import (
	"fmt"
	"github.com/blang/semver"
	"syscall"
	"github.com/springernature/halfpipe/sync"
	"github.com/springernature/halfpipe/sync/githubRelease"
	"os"
)

var (
	// This field will be populated in Concourse from the version resource
	// go build -ldflags "-X main.version`cat version/version`"
	version string
)

func getVersion() (semver.Version, error) {
	if version == "" {
		version, _ := semver.Make("0.0.0-DEV")
		return version, nil
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
	currentVersion, err := getVersion()
	printAndExit(err)

	sync := sync.Syncer{CurrentVersion: currentVersion, GithubRelease: githubRelease.GithubRelease{}}
	if len(os.Args) == 1 {
		printAndExit(sync.Check())
	} else if len(os.Args) > 1 && os.Args[1] == "sync" {
		printAndExit(sync.Update())
		return
	}

	fmt.Println("Hello World")
	fmt.Println("Current version is: " + currentVersion.String())
}
