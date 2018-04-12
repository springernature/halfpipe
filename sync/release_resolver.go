package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

type results struct {
	Results []result `json:"results"`
}

func (r results) GetLatest() (result Release) {
	highestVersion := semver.Version{}

	for _, res := range r.Results {
		currVersion, err := res.getVersion()
		if err != nil {
			// Means the binary name is wrong, just skip it
			continue
		}

		if currVersion.GT(highestVersion) {
			highestVersion = currVersion
			result = Release{
				Version:     currVersion,
				DownloadURL: res.getDownloadURL(),
			}
		}
	}

	return
}

type result struct {
	URI string `json:"uri"`
}

func (r result) getDownloadURL() string {
	return strings.Replace(r.URI, "api/storage/", "", -1)
}

func (r result) getVersion() (version semver.Version, err error) {
	rx := regexp.MustCompile(`[1-9]+.[1-9]+.[1-9]+$`)
	version, err = semver.Parse(string(rx.Find([]byte(r.URI))))
	return
}

type Release struct {
	Version     semver.Version
	DownloadURL string
}

type HTTPGetter func(url string) (resp *http.Response, err error)

var ResolveLatestVersionFromArtifactory = func(os string, httpGetter HTTPGetter) (release Release, err error) {
	url := fmt.Sprintf("https://springernature.jfrog.io/springernature/api/search/artifact?name=halfpipe_%s", os)
	resp, err := httpGetter(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var r results
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}

	release = r.GetLatest()
	return
}
