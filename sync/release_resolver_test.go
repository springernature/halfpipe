package sync

import (
	"errors"
	"net/http"
	"testing"

	"github.com/blang/semver"
	"github.com/onsi/gomega/gbytes"
	"github.com/stretchr/testify/assert"
)

func TestReleaseResolverCallsOutToTheCorrectUrl(t *testing.T) {
	var actualUrl string
	fakeHttpGetter := func(url string) (resp *http.Response, err error) {
		actualUrl = url
		resp = &http.Response{
			Body: gbytes.BufferWithBytes([]byte("")),
		}
		return
	}

	ResolveLatestVersionFromArtifactory("darwin", fakeHttpGetter)

	expected := "https://springernature.jfrog.io/springernature/api/search/artifact?name=halfpipe_darwin"
	assert.Equal(t, expected, actualUrl)
}

func TestReleaseResolverReturnsTheErrorFromHttpGetter(t *testing.T) {
	exptectedError := errors.New("Blurgh")
	fakeHttpGetter := func(url string) (resp *http.Response, err error) {
		err = exptectedError
		return
	}

	_, err := ResolveLatestVersionFromArtifactory("darwin", fakeHttpGetter)

	assert.Equal(t, exptectedError, err)
}

func TestGivesTheCorrectRelease(t *testing.T) {
	returnFromHttpCall := `
{
 "results" : [
	{"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/somethingRandom"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin"},
	{"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin_1.21.7"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin_1.21.6"}
 ]
}
`

	fakeHttpGetter := func(url string) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: gbytes.BufferWithBytes([]byte(returnFromHttpCall)),
		}
		return
	}

	release, err := ResolveLatestVersionFromArtifactory("darwin", fakeHttpGetter)

	expectedResults := Release{
		Version:     semver.Version{Major: 1, Minor: 21, Patch: 7},
		DownloadURL: "https://springernature.jfrog.io/springernature/halfpipe/halfpipe_darwin_1.21.7",
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedResults, release)
}
