package upload

import (
	"io"
	"os/exec"
	"testing"

	"github.com/concourse/fly/rc"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var fs = afero.Afero{Fs: afero.NewMemMapFs()}
var currentDir = "/blurg"
var stdout io.Writer
var stderr io.Writer
var stdin io.Reader

var validManifest = manifest.Manifest{
	Team:     "my-team",
	Pipeline: "my-pipeline",
}

var validManifestReader = func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
	man = validManifest
	return
}

var lookPath = func(path string) (string, error) {
	return path, nil
}

func TestPassesOnErrorFromManifestReader(t *testing.T) {
	expectedError := errors.New("meehp")

	targetsReader := func() (target Targets, err error) {
		return
	}

	manifestReader := func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		err = expectedError
		return
	}

	upload := NewUpload(targetsReader, manifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	_, err := upload.CreatePlan()

	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfPipelineOrTeamIsEmpty(t *testing.T) {
	targetsReader := func() (target Targets, err error) {
		return
	}

	manifestReader := func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		man = manifest.Manifest{
			Team: "",
		}
		return
	}

	upload := NewUpload(targetsReader, manifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	_, err := upload.CreatePlan()

	assert.Error(t, err)

	manifestReader = func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		man = manifest.Manifest{
			Team: "Kehe",
		}
		return
	}

	upload = NewUpload(targetsReader, manifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	_, err = upload.CreatePlan()

	assert.Error(t, err)

}

func TestPassesOnErrorFromTargetsReader(t *testing.T) {
	expectedError := errors.New("meehp")

	targetsReader := func() (target Targets, err error) {
		err = expectedError
		return
	}

	upload := NewUpload(targetsReader, validManifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	_, err := upload.CreatePlan()

	assert.Equal(t, expectedError, err)
}

func TestCreatesALoginPlanIfTargetIsEmpty(t *testing.T) {
	targetsReader := func() (target Targets, err error) {
		target = map[rc.TargetName]rc.TargetProps{}
		return
	}

	upload := NewUpload(targetsReader, validManifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	plan, err := upload.CreatePlan()

	assert.Nil(t, err)
	expectedCommand := Command{
		Cmd: &exec.Cmd{
			Path:   "fly",
			Args:   []string{"fly", "-t", validManifest.Team, "login", "-c", "https://concourse.halfpipe.io", "-n", validManifest.Team},
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		Printable: "fly -t my-team login -c https://concourse.halfpipe.io -n my-team",
	}
	assert.Equal(t, expectedCommand, plan[0])
}

func TestCreatesALoginPlanIfTargetDoesNotContainTeam(t *testing.T) {
	targetsReader := func() (target Targets, err error) {
		target = map[rc.TargetName]rc.TargetProps{
			"someTeam": {},
		}
		return
	}

	upload := NewUpload(targetsReader, validManifestReader, stdout, stderr, stdin, lookPath, fs, currentDir)
	plan, err := upload.CreatePlan()

	assert.Nil(t, err)

	expectedCommand := Command{
		Cmd: &exec.Cmd{
			Path:   "fly",
			Args:   []string{"fly", "-t", validManifest.Team, "login", "-c", "https://concourse.halfpipe.io", "-n", validManifest.Team},
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		Printable: "fly -t my-team login -c https://concourse.halfpipe.io -n my-team",
	}

	assert.Equal(t, expectedCommand.Cmd, plan[0].Cmd)
}
