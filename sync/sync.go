package sync

import (
	"github.com/springernature/halfpipe/sync/githubRelease"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"net/http"
	"strconv"
	"gopkg.in/cheggaaa/pb.v1"
	"github.com/springernature/halfpipe"
)

type Sync interface {
	Check() error
	Update() error
}
type Syncer struct {
	CurrentVersion semver.Version
	GithubRelease  githubRelease.GithubReleaser
}

func (s Syncer) Check() error {
	if s.CurrentVersion.EQ(halfpipe.DevVersion) {
		return nil
	}

	latestVersion, err := s.GithubRelease.GetLatestVersion()
	if err != nil {
		return err
	}

	if latestVersion.GT(s.CurrentVersion) {
		errorMessage := fmt.Sprintf("Current version %s is behind latest version %s. Please run 'halfpipe sync'", s.CurrentVersion.String(), latestVersion.String())
		return errors.New(errorMessage)
	}
	return nil
}

func (s Syncer) Update() error {
	if s.CurrentVersion.EQ(halfpipe.DevVersion) {
		return errors.New("Can not upgrade dev release...")
	}

	url, err := s.GithubRelease.GetLatestBinaryURL()
	if err != nil {
		return err
	}

	updateOptions := update.Options{}
	err = updateOptions.CheckPermissions()
	if err != nil {
		return err
	}
	fmt.Printf("downloading latest version from %s... \n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	filesSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	progressBar := pb.New64(filesSize).SetUnits(pb.U_BYTES)
	progressBar.Start()
	defer progressBar.FinishPrint(fmt.Sprintf("successfully updated"))
	reader := progressBar.NewProxyReader(resp.Body)

	err = update.Apply(reader, updateOptions)
	if err != nil {
		return err
	}

	return nil
}
