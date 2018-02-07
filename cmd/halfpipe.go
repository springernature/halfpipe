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
	fmt.Println("Hello World")

	sync := sync.Sync{CurrentVersion: currentVersion, GithubRelease: githubRelease.GithubRelease{}}
	fmt.Println(os.Args)
	if len(os.Args) == 1 {
		err = sync.Check()
		printAndExit(err)
	} else if len(os.Args) > 1 && os.Args[1] == "sync" {
		err = sync.Update()
		printAndExit(err)
		fmt.Println("Yay, updated binary!")
		syscall.Exit(0)
	}

	fmt.Println("Current version is: " + currentVersion.String())
}
