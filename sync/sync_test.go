package sync

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/inconshreveable/go-update"
	"github.com/onsi/gomega/gbytes"
	"github.com/stretchr/testify/assert"
)

type ReleaseResolverDouble struct {
	getLatestRelease func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

func (r ReleaseResolverDouble) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	return r.getLatestRelease(ctx, owner, repo)
}

func TestDoesNothingWhenCurrentVersionIsDevRelease(t *testing.T) {
	release := sync{
		currentVersion: DevVersion,
	}
	err := release.Check()
	assert.Nil(t, err)
}

func TestCheckReturnsNilWhenCurrentVersionIsUpToDate(t *testing.T) {

	latestVersion := semver.Version{Major: 1}

	release := sync{
		currentVersion: latestVersion,
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				tagName := latestVersion.String()
				release := github.RepositoryRelease{
					TagName: &tagName,
				}
				return &release, nil, nil
			},
		},
	}

	err := release.Check()
	assert.Nil(t, err)

}

func TestPassesOnErrorFromReleaseResolver(t *testing.T) {
	releaseError := errors.New("Noooes")

	release := sync{
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				return nil, nil, releaseError
			},
		},
	}

	err := release.Check()
	assert.Equal(t, releaseError, err)

}

func TestCheckReturnsErrorWhenWeCannotParseTheTagFromTheRelease(t *testing.T) {
	release := sync{
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				tagName := "MyCoolTag"
				release := github.RepositoryRelease{
					TagName: &tagName,
				}
				return &release, nil, nil
			},
		},
	}

	err := release.Check()
	assert.Error(t, err)

}

func TestCheckReturnsErrorWhenCurrentVersionIsBehind(t *testing.T) {

	currentVersion := semver.Version{Major: 0}
	latestVersion := semver.Version{Major: 1}

	release := sync{
		currentVersion: currentVersion,
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				tagName := latestVersion.String()
				release := github.RepositoryRelease{
					TagName: &tagName,
				}
				return &release, nil, nil
			},
		},
	}

	err := release.Check()
	assert.Error(t, err)
}

func TestUpdateErrorsOutIfTryingToUpdateDevRelease(t *testing.T) {
	release := sync{
		currentVersion: DevVersion,
	}
	err := release.Update(&bytes.Buffer{})
	assert.Error(t, err)
}

func TestUpdateErrorsOutIfWeCannotGetLatestRelease(t *testing.T) {
	releaseError := errors.New("asd")
	release := sync{
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				return nil, nil, releaseError
			},
		},
	}
	err := release.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, releaseError)
}

func TestUpdateErrorsOutIfWeCannotFindDownloadUrlForOurArch(t *testing.T) {
	release := sync{
		os: "windows",
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				downloadUrlOsx := "https:///blablabla/binary-osx"
				downloadUrlLinux := "https:///blablabla/binary-linux"
				release := github.RepositoryRelease{
					Assets: []github.ReleaseAsset{
						{BrowserDownloadURL: &downloadUrlLinux},
						{BrowserDownloadURL: &downloadUrlOsx},
					},
				}
				return &release, nil, nil
			},
		},
	}

	err := release.Update(&bytes.Buffer{})
	assert.Error(t, err)

}

func TestUpdateErrorsOutIfWeFailToDownload(t *testing.T) {
	httpError := errors.New("Shiet")
	release := sync{
		os: "osx",
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				downloadUrlOsx := "https:///blablabla/binary-osx"
				release := github.RepositoryRelease{
					Assets: []github.ReleaseAsset{
						{BrowserDownloadURL: &downloadUrlOsx},
					},
				}
				return &release, nil, nil
			},
		},
		httpGetter: func(url string) (resp *http.Response, err error) {
			err = httpError
			return
		},
	}

	err := release.Update(&bytes.Buffer{})
	assert.Error(t, err)
	assert.Equal(t, err, httpError)

}

func TestUpdateReturnsUpdateErrorFromUpdater(t *testing.T) {
	downloadUrlOsx := "https:///blablabla/binary-osx"
	updateError := errors.New("Buuh")

	release := sync{
		os: "osx",
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
				release := github.RepositoryRelease{
					Assets: []github.ReleaseAsset{
						{BrowserDownloadURL: &downloadUrlOsx},
					},
				}
				return &release, nil, nil
			},
		},
		httpGetter: func(url string) (resp *http.Response, err error) {
			resp = &http.Response{}
			return
		},
		updater: func(update io.Reader, opts update.Options) error { return updateError },
	}

	err := release.Update(&bytes.Buffer{})
	assert.Equal(t, err, updateError)

}

func TestUpdateDoesWhatItShouldDo(t *testing.T) {
	downloadUrlOsx := "https:///blablabla/binary-osx"

	var calledOutToHttpGetter bool
	var calledOutToUpdater bool
	release := sync{
		os: "osx",
		releaseResolver: ReleaseResolverDouble{
			func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {

				release := github.RepositoryRelease{
					Assets: []github.ReleaseAsset{
						{BrowserDownloadURL: &downloadUrlOsx},
					},
				}
				return &release, nil, nil
			},
		},
		httpGetter: func(url string) (resp *http.Response, err error) {
			calledOutToHttpGetter = true

			resp = &http.Response{
				Body: gbytes.NewBuffer(),
			}
			return
		},
		updater: func(update io.Reader, opts update.Options) error {
			calledOutToUpdater = true
			return nil
		},
	}

	err := release.Update(&bytes.Buffer{})
	assert.Nil(t, err)
	assert.True(t, calledOutToHttpGetter)
	assert.True(t, calledOutToUpdater)
}
