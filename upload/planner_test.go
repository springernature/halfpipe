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
var branch = "my-branch"
var validPipeline = fmt.Sprintf(`---
team: %s
pipeline: %s
`, team, pipeline)

var validPipelineWithBranch = fmt.Sprintf(`---
team: %s
pipeline: %s
repo:
  branch: %s
`, team, pipeline, branch)

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

var NullpipelineFile = func(fs afero.Afero) (file afero.File, err error) {
	return
}

func TestReturnsErrWhenHalfpipeFileDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, NullpipelineFile, false)
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsErrWhenHalfpipeDoesntContainTeamOrPipeline(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(""), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, NullpipelineFile, false)
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsAPlanWithLogin(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false)
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "fly",
				Args:   []string{"fly", "-t", team, "login", "-c", "https://concourse.halfpipe.io", "-n", team},
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  stdin,
			},
		},
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stderr: stderr,
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path:   "fly",
				Args:   []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml"},
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  stdin,
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAPlanWithoutLoginIfAlreadyLoggedIn(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false)
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stderr: stderr,
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path:   "fly",
				Args:   []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml"},
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  stdin,
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAPlanWithoutLoginIfAlreadyLoggedInAndWithBranch(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipelineWithBranch), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false)
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stderr: stderr,
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path:   "fly",
				Args:   []string{"fly", "-t", team, "set-pipeline", "-p", pipeline + "-" + branch, "-c", "pipeline.yml"},
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  stdin,
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAPlanWithNonInteractiveIfSpecified(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, stdout, stderr, stdin, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, true)
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stderr: stderr,
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path:   "fly",
				Args:   []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--non-interactive"},
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  stdin,
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}
