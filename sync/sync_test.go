package sync

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"io"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/inconshreveable/go-update"
	"github.com/onsi/gomega/gbytes"
	"github.com/springernature/halfpipe/config"
	"github.com/stretchr/testify/assert"
)

type releaseResolverDouble struct {
	getLatestRelease func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)

	tagName         string
	err             error
	releaseAssetURL []string
}

func newReleaseResolverDouble() releaseResolverDouble {
	return releaseResolverDouble{}
}

func (r releaseResolverDouble) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	var releaseAssets []github.ReleaseAsset
	for _, url := range r.releaseAssetURL {
		releaseAssets = append(releaseAssets, github.ReleaseAsset{BrowserDownloadURL: &url})
	}

	release := github.RepositoryRelease{
		TagName: &r.tagName,
		Assets:  releaseAssets,
	}
	return &release, nil, r.err
}

func (r releaseResolverDouble) SetLatestReleaseVersion(tagName string) releaseResolverDouble {
	r.tagName = tagName
	return r
}

func (r releaseResolverDouble) SetError(err error) releaseResolverDouble {
	r.err = err
	return r
}

func (r releaseResolverDouble) AddReleaseAssetURL(url string) releaseResolverDouble {
	r.releaseAssetURL = append(r.releaseAssetURL, url)
	return r
}

func TestDoesNothingWhenCurrentVersionIsDevRelease(t *testing.T) {
	release := sync{
		currentVersion: config.DevVersion,
	}
	err := release.Check()
	assert.Nil(t, err)
}

func TestCheckReturnsNilWhenCurrentVersionIsUpToDate(t *testing.T) {

	latestVersion := semver.Version{Major: 1}

	syncer := NewSyncer(latestVersion, newReleaseResolverDouble().SetLatestReleaseVersion(latestVersion.String()))

	err := syncer.Check()
	assert.Nil(t, err)

}

func TestPassesOnErrorFromReleaseResolver(t *testing.T) {
	releaseError := errors.New("Noooes")

	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().SetError(releaseError))

	err := syncer.Check()
	assert.Equal(t, releaseError, err)

}

func TestCheckReturnsErrorWhenWeCannotParseTheTagFromTheRelease(t *testing.T) {
	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().SetLatestReleaseVersion("MyCoolTag"))

	err := syncer.Check()
	assert.Error(t, err)
}

func TestCheckReturnsErrorWhenCurrentVersionIsBehind(t *testing.T) {

	currentVersion := semver.Version{}
	latestVersion := semver.Version{Major: 1}

	syncer := NewSyncer(currentVersion, newReleaseResolverDouble().SetLatestReleaseVersion(latestVersion.String()))

	err := syncer.Check()
	assert.Error(t, err)
	assert.Equal(t, err, ErrOutOfDateBinary(currentVersion, latestVersion))
}

func TestUpdateErrorsOutIfTryingToUpdateDevRelease(t *testing.T) {
	syncer := NewSyncer(config.DevVersion, newReleaseResolverDouble())

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, ErrUpdatingDevRelease)
}

func TestUpdateErrorsOutIfWeCannotGetLatestRelease(t *testing.T) {
	releaseError := errors.New("asd")

	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().SetError(releaseError))

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, releaseError)
}

func TestUpdateErrorsOutIfWeCannotFindDownloadUrlForOurArch(t *testing.T) {
	syncer := NewSyncer(semver.Version{Major: 1},
		newReleaseResolverDouble().
			AddReleaseAssetURL("https:///blablabla/binary-osx").
			AddReleaseAssetURL("https:///blablabla/binary-linux"))
	syncer.os = "windows"

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, ErrNoBinaryForArch(syncer.os))
}

func TestUpdateErrorsOutIfWeFailToDownload(t *testing.T) {
	httpError := errors.New("Shiet")
	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().AddReleaseAssetURL("https:///blablabla/binary-osx"))
	syncer.os = "osx"
	syncer.httpGetter = func(url string) (resp *http.Response, err error) {
		err = httpError
		return
	}

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, httpError)
}

func TestUpdateReturnsUpdateErrorFromUpdater(t *testing.T) {
	updateError := errors.New("Buuh")

	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().SetError(updateError))
	syncer.os = "osx"
	syncer.httpGetter = func(url string) (resp *http.Response, err error) {
		resp = &http.Response{}
		return
	}

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, updateError)
}

func TestUpdateDoesWhatItShouldDo(t *testing.T) {
	syncer := NewSyncer(semver.Version{Major: 1}, newReleaseResolverDouble().AddReleaseAssetURL("https:///blablabla/binary-osx"))
	syncer.os = "osx"

	var calledOutToHTTPGetter bool
	syncer.httpGetter = func(url string) (resp *http.Response, err error) {
		calledOutToHTTPGetter = true

		resp = &http.Response{
			Body: gbytes.NewBuffer(),
		}
		return
	}

	var calledOutToUpdater bool
	syncer.updater = func(update io.Reader, opts update.Options) error {
		calledOutToUpdater = true
		return nil
	}

	err := syncer.Update(&bytes.Buffer{})
	assert.Nil(t, err)
	assert.True(t, calledOutToHTTPGetter)
	assert.True(t, calledOutToUpdater)
}
