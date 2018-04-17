package upload

import (
	"fmt"
	"os/exec"
	"path"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var team = "my-team"
var pipeline = "my-pipeline"
var validPipeline = fmt.Sprintf(`---
team: %s
pipeline: %s
`, team, pipeline)

var validFlyRc = fmt.Sprintf(`---
targets:
  %s: {}`, team)

var homedir = "/home/my-user"

var stdin = gbytes.NewBuffer()
var stdout = gbytes.NewBuffer()
var stderr = gbytes.NewBuffer()

var pathResolver = func(path string) (string, error) {
	return path, nil
}

func TestReturnsErrWhenHalfpipeFileDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin)
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsErrWhenHalfpipeDoesntContainTeamOrPipeline(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(""), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin)
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsAPlanWithLogin(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin)
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{exec.Cmd{
			Path:   "fly",
			Args:   []string{"fly", "-t", team, "login", "-c", "https://concourse.halfpipe.io", "-n", team},
			Stdout: stdout,
			Stderr: stderr,
			Stdin:  stdin,
		}},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAPlanWithoutLoginIfAlreadyLoggedIn(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin)
	plan, err := planner.Plan()

	//expectedPlan := nil

	assert.Nil(t, err)
	assert.Nil(t, plan)
}
