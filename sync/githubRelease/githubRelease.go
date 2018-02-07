package githubRelease

import (
	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"context"
)

type GithubReleaseI interface {
	GetLatestVersion() (semver.Version, error)
	GetLatestBinaryURLS() (assets []string, err error)
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

func (GithubRelease) GetLatestBinaryURLS() (assets []string, err error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "springernature", "halfpipe")
	if err != nil {
		return assets, err
	}

	for _, asset := range release.Assets {
		assets = append(assets, *asset.BrowserDownloadURL)
	}
	return assets, nil
}
