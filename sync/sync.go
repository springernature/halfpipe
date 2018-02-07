package sync

import (
	"github.com/springernature/halfpipe/sync/githubRelease"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"fmt"
	"runtime"
	"strings"
	"github.com/inconshreveable/go-update"
	"net/http"
)

type Sync struct {
	CurrentVersion semver.Version
	GithubRelease  githubRelease.GithubReleaseI
}

func (s Sync) Check() error {
	latestVersion, err := s.GithubRelease.GetLatestVersion()
	if err != nil {
		return err
	}

	if latestVersion.GT(s.CurrentVersion) {
		errorMessage := fmt.Sprintf("Current version %s is behind latest version %s. Please run 'halfpipe sync'", s.CurrentVersion.String(), latestVersion.String())
		return errors.New(errorMessage)
	}
	return nil
}

func (s Sync) getDownloadUrl() (string, error) {
	url, err := s.GithubRelease.GetLatestBinaryURLS()
	if err != nil {
		return "", err
	}
	for _, u := range url {
		if strings.Contains(u, runtime.GOOS) {
			return u, nil
		}
	}
	return "", errors.New("Could not find a binary for your OS..")
}

func (s Sync) Update() error {
	url, err := s.getDownloadUrl()
	if err != nil {
		return err
	}
	fmt.Printf("downloading halfpipe from %s... \n", url)

	updateOptions := update.Options{}
	err = updateOptions.CheckPermissions()
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	err = update.Apply(resp.Body, updateOptions)
	if err != nil {
		return err
	}

	return nil
}
