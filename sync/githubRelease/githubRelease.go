package githubRelease

import (
	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"context"
	"strings"
	"runtime"
	"errors"
)

type GithubReleaser interface {
	GetLatestVersion() (semver.Version, error)
	GetLatestBinaryURL() (string, error)
}

type GithubRelease struct{}

func (GithubRelease) GetLatestVersion() (semver.Version, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "springernature", "halfpipe")
	if err != nil {
		return semver.Version{}, err
	}
	version, err := semver.Make(*release.TagName)
	if err != nil {
		return semver.Version{}, err
	}
	return version, nil
}

func (GithubRelease) GetLatestBinaryURL() (string, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "springernature", "halfpipe")
	if err != nil {
		return "", err
	}

	for _, asset := range release.Assets {
		if strings.Contains(*asset.BrowserDownloadURL, runtime.GOOS) {
			downloadUrl := *asset.BrowserDownloadURL
			return downloadUrl, nil
		}
	}
	return "", errors.New("Could not find a binary for your OS..")
}
