package sync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/inconshreveable/go-update"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	NoBinaryForArchError = func(os string) error {
		return errors.New(fmt.Sprintf("Could not find a binary for your arch, '%s'", os))
	}
	UpdatingDevReleaseError = errors.New("cannot update a dev release")
	OutOfDateBinaryError    = func(currentVersion semver.Version, latestVersion semver.Version) error {
		errorMessage := fmt.Sprintf("Current version %s is behind latest version %s. Please run 'halfpipe sync'", currentVersion, latestVersion)
		return errors.New(errorMessage)
	}
)

type Sync interface {
	Check() error
	Update(out io.Writer) error
}

type LatestReleaseResolver interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

type sync struct {
	currentVersion  semver.Version
	releaseResolver LatestReleaseResolver
	os              string

	httpGetter func(url string) (resp *http.Response, err error)
	updater    func(update io.Reader, opts update.Options) error
}

func NewSyncer(currentRelease semver.Version) sync {
	return sync{
		currentVersion:  currentRelease,
		releaseResolver: github.NewClient(nil).Repositories,
		os:              runtime.GOOS,
		httpGetter:      http.Get,
		updater:         update.Apply,
	}
}

func (s sync) getLatestRelease() (release *github.RepositoryRelease, err error) {
	release, _, err = s.releaseResolver.GetLatestRelease(context.Background(), "springernature", "halfpipe")
	return
}

func (s sync) Check() (err error) {
	if s.currentVersion.EQ(DevVersion) {
		return
	}

	latestRelease, err := s.getLatestRelease()
	if err != nil {
		return
	}

	latestVersion, err := semver.Parse(*latestRelease.TagName)
	if err != nil {
		return
	}

	if s.currentVersion.LT(latestVersion) {
		err = OutOfDateBinaryError(s.currentVersion, latestVersion)
	}

	return
}

func (s sync) getLatestBinaryUrl() (url string, err error) {
	release, err := s.getLatestRelease()
	if err != nil {
		return
	}

	for _, asset := range release.Assets {
		if strings.Contains(*asset.BrowserDownloadURL, s.os) {
			url = *asset.BrowserDownloadURL
			return
		}
	}
	err = NoBinaryForArchError(s.os)
	return
}

func (s sync) Update(out io.Writer) (err error) {
	if s.currentVersion.EQ(DevVersion) {
		return UpdatingDevReleaseError
	}

	binaryUrl, err := s.getLatestBinaryUrl()
	if err != nil {
		return
	}

	out.Write([]byte("downloading latest version from " + binaryUrl + "\n"))
	resp, err := s.httpGetter(binaryUrl)
	if err != nil {
		return
	}

	filesSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	progressBar := pb.New64(filesSize).SetUnits(pb.U_BYTES)
	progressBar.Output = out
	progressBar.Start()
	defer progressBar.FinishPrint(fmt.Sprintf("successfully updated"))

	err = s.updater(progressBar.NewProxyReader(resp.Body), update.Options{})

	return
}
