package sync

import (
	"testing"

	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/blang/semver"
	"github.com/inconshreveable/go-update"
	"github.com/onsi/gomega/gbytes"
	"github.com/springernature/halfpipe/config"
	"github.com/stretchr/testify/assert"
)

func releaseResolverDouble(r Release, e error) LatestReleaseResolver {
	return func(os string, httpGetter HttpGetter) (release Release, err error) {
		return r, e
	}
}

func TestDoesNothingWhenCurrentVersionIsDevRelease(t *testing.T) {
	release := sync{
		currentVersion: config.DevVersion,
	}
	err := release.Check()
	assert.Nil(t, err)
}

func TestDoesNothingWhenCheckShouldBeSkipped(t *testing.T) {
	release := sync{
		currentVersion: semver.Version{Major: 1},
		shouldSkip:     true,
	}

	err := release.Check()
	assert.Nil(t, err)
}

func TestCheckReturnsNilWhenCurrentVersionIsUpToDate(t *testing.T) {

	latestVersion := semver.Version{Major: 1}
	latestRelease := Release{
		Version: latestVersion,
	}

	syncer := NewSyncer(latestVersion, releaseResolverDouble(latestRelease, nil))

	err := syncer.Check()
	assert.Nil(t, err)

}

func TestPassesOnErrorFromReleaseResolver(t *testing.T) {
	releaseError := errors.New("Noooes")

	syncer := NewSyncer(semver.Version{Major: 1}, releaseResolverDouble(Release{}, releaseError))

	err := syncer.Check()
	assert.Equal(t, releaseError, err)
}

func TestCheckReturnsErrorWhenCurrentVersionIsBehind(t *testing.T) {

	currentVersion := semver.Version{}
	latestVersion := semver.Version{Major: 1}

	syncer := NewSyncer(currentVersion, releaseResolverDouble(Release{Version: latestVersion}, nil))

	err := syncer.Check()
	assert.Error(t, err)
	assert.Equal(t, err, ErrOutOfDateBinary(currentVersion, latestVersion))
}

func TestUpdateErrorsOutIfTryingToUpdateDevRelease(t *testing.T) {
	syncer := NewSyncer(config.DevVersion, releaseResolverDouble(Release{}, nil))

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, ErrUpdatingDevRelease)
}

func TestUpdateErrorsOutIfWeFailToDownload(t *testing.T) {
	httpError := errors.New("Shiet")
	syncer := NewSyncer(semver.Version{Major: 1}, releaseResolverDouble(Release{Version: semver.Version{Major: 2}}, nil))

	syncer.os = "osx"
	syncer.httpGetter = func(url string) (resp *http.Response, err error) {
		err = httpError
		return
	}

	err := syncer.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, httpError)
}

func TestUpdateDoesWhatItShouldDo(t *testing.T) {
	syncer := NewSyncer(semver.Version{Major: 1}, releaseResolverDouble(Release{Version: semver.Version{Major: 2}}, nil))

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
