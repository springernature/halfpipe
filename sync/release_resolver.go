package sync

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/blang/semver"
	"regexp"
	"strings"
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
				Version: currVersion,
				DownloadURL: res.getDownloadUrl(),
			}
		}
	}

	return
}

type result struct {
	Uri string `json:"uri"`
}

func (r result) getDownloadUrl() string {
	return strings.Replace(r.Uri, "api/storage/", "", -1)
}

func (r result) getVersion() (version semver.Version, err error) {
	rx := regexp.MustCompile(`[1-9]+.[1-9]+.[1-9]+$`)
	version, err = semver.Parse(string(rx.Find([]byte(r.Uri))))
	return
}

type Release struct {
	Version     semver.Version
	DownloadURL string
}

type HttpGetter func(url string) (resp *http.Response, err error)

var ResolveLatestVersionFromArtifactory = func(os string, httpGetter HttpGetter) (release Release, err error) {
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
	json.Unmarshal(body, &r)

	release = r.GetLatest()
	return
}
