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
	var actualURL string
	fakeHTTPGetter := func(url string) (resp *http.Response, err error) {
		actualURL = url
		resp = &http.Response{
			Body: gbytes.BufferWithBytes([]byte("")),
		}
		return
	}

	ResolveLatestVersionFromArtifactory("darwin", fakeHTTPGetter)

	expected := "https://springernature.jfrog.io/springernature/api/search/artifact?name=halfpipe_darwin"
	assert.Equal(t, expected, actualURL)
}

func TestReleaseResolverReturnsTheErrorFromHttpGetter(t *testing.T) {
	exptectedError := errors.New("Blurgh")
	fakeHTTPGetter := func(url string) (resp *http.Response, err error) {
		err = exptectedError
		return
	}

	_, err := ResolveLatestVersionFromArtifactory("darwin", fakeHTTPGetter)

	assert.Equal(t, exptectedError, err)
}

func TestGivesTheCorrectRelease(t *testing.T) {
	returnFromHTTPCall := `
{
 "results" : [
	{"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/somethingRandom"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin"},
	{"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin_1.22.0"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe"},
    {"uri" : "https://springernature.jfrog.io/springernature/api/storage/halfpipe/halfpipe_darwin_1.21.6"}
 ]
}
`

	fakeHTTPGetter := func(url string) (resp *http.Response, err error) {
		resp = &http.Response{
			Body: gbytes.BufferWithBytes([]byte(returnFromHTTPCall)),
		}
		return
	}

	release, err := ResolveLatestVersionFromArtifactory("darwin", fakeHTTPGetter)

	expectedResults := Release{
		Version:     semver.Version{Major: 1, Minor: 22, Patch: 0},
		DownloadURL: "https://springernature.jfrog.io/springernature/halfpipe/halfpipe_darwin_1.22.0",
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedResults, release)
}
