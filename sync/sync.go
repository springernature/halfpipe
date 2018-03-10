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
	"github.com/springernature/halfpipe/config"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	ErrNoBinaryForArch = func(os string) error {
		return fmt.Errorf("could not find a binary for your arch, '%s'", os)
	}
	ErrUpdatingDevRelease = errors.New("cannot update a dev release")
	ErrOutOfDateBinary    = func(currentVersion semver.Version, latestVersion semver.Version) error {
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

func NewSyncer(currentRelease semver.Version, releaseResolver LatestReleaseResolver) sync {
	return sync{
		currentVersion:  currentRelease,
		releaseResolver: releaseResolver,
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
	if s.currentVersion.EQ(config.DevVersion) {
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
		err = ErrOutOfDateBinary(s.currentVersion, latestVersion)
	}

	return
}

func (s sync) getLatestBinaryURL() (url string, err error) {
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
	err = ErrNoBinaryForArch(s.os)
	return
}

func (s sync) Update(out io.Writer) (err error) {
	if s.currentVersion.EQ(config.DevVersion) {
		return ErrUpdatingDevRelease
	}

	binaryURL, err := s.getLatestBinaryURL()
	if err != nil {
		return
	}

	_, err = out.Write([]byte("downloading latest version from " + binaryURL + "\n"))
	if err != nil {
		return
	}

	resp, err := s.httpGetter(binaryURL)
	if err != nil {
		return
	}

	var fileSize int64
	fileSize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		fileSize = 15000000 // just guess for the progress bar
	}

	progressBar := pb.New64(fileSize).SetUnits(pb.U_BYTES)
	progressBar.Output = out
	progressBar.Start()
	defer progressBar.FinishPrint(fmt.Sprintf("successfully updated"))

	err = s.updater(progressBar.NewProxyReader(resp.Body), update.Options{})
	return
}
