package upload

import (
	"fmt"
	"os/exec"
	"path"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/pkg/errors"
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
  %s:
    api: https://concourse.domain.io`, team)

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

var osResolver = func() string {
	return "darwin"
}

var envResolver = func(envVar string) string {
	return ""
}

func TestReturnsErrWhenHalfpipeFileDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	planner := NewPlanner(fs, pathResolver, homedir, NullpipelineFile, false, "master", osResolver, envResolver, "")
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsReadableErrorWhenFlyIsNotOnPath(t *testing.T) {
	pathResolverWithoutFly := func(path string) (string, error) {
		if path == "fly" {
			return "", errors.New("Some random error")
		}
		return path, nil
	}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)

	planner := NewPlanner(fs, pathResolverWithoutFly, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "master", osResolver, envResolver, "")

	_, err := planner.Plan()
	assert.Equal(t, ErrFlyNotInstalled(osResolver()), err)
}

func TestReturnsErrWhenHalfpipeDoesntContainTeamOrPipeline(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(""), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, NullpipelineFile, false, "master", osResolver, envResolver, "")
	_, err := planner.Plan()

	assert.Error(t, err)
}

func TestReturnsAPlanWithLogin(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "master", osResolver, envResolver, "")
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "login", "-c", "https://concourse.halfpipe.io", "-n", team},
			},
		},
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds"},
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

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "master", osResolver, envResolver, "")
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds"},
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

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "master", osResolver, envResolver, "")
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline + "-" + branch, "-c", "pipeline.yml", "--check-creds"},
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

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, true, "master", osResolver, envResolver, "")
	plan, err := planner.Plan()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path:   "halfpipe",
				Args:   []string{"halfpipe"},
				Stdout: file,
			},
			Printable: "halfpipe > pipeline.yml",
		},
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "set-pipeline", "-p", pipeline, "-c", "pipeline.yml", "--check-creds", "--non-interactive"},
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAUnpausePlan(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, true, "master", osResolver, envResolver, "")
	plan, err := planner.Unpause()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "unpause-pipeline", "-p", pipeline},
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAUnpausePlanForBranch(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipelineWithBranch), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, true, "master", osResolver, envResolver, "")
	plan, err := planner.Unpause()

	expectedPlan := Plan{
		{
			Cmd: exec.Cmd{
				Path: "fly",
				Args: []string{"fly", "-t", team, "unpause-pipeline", "-p", fmt.Sprintf("%s-%s", pipeline, branch)},
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, plan)
}

func TestReturnsAPlanWithSecurityQuestionIfNotOnMaster(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "a-branch", osResolver, envResolver, "")
	plan, err := planner.Plan()

	assert.NoError(t, err)
	assert.Len(t, plan, 3)
	assert.Equal(t, "# Security question", plan[0].Printable)
}

func TestReturnsAPlanWithoutSecurityQuestionIfNotOnMasterAndNonInteractiveSet(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, true, "a-branch", osResolver, envResolver, "")
	plan, err := planner.Plan()

	assert.NoError(t, err)
	assert.Len(t, plan, 2)
	assert.NotEqual(t, "# Security question", plan[0].Printable)
	assert.NotEqual(t, "# Security question", plan[1].Printable)
}

func TestCanOverrideTheApi(t *testing.T) {
	concourseEndpoint := "https://someRandom.location.com"
	envResolver := func(envVar string) string {
		return concourseEndpoint
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	file, _ := fs.Create("pipeline.yml")
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	fs.WriteFile(path.Join(homedir, ".flyrc"), []byte(validFlyRc), 0777)

	planner := NewPlanner(fs, pathResolver, homedir, func(fs afero.Afero) (afero.File, error) {
		return file, nil
	}, false, "master", osResolver, envResolver, "")
	plan, err := planner.Plan()

	assert.NoError(t, err)
	fmt.Println(plan)

	assert.Equal(t, plan[1].Cmd.Args, []string{"fly", "-t", team, "login", "-c", concourseEndpoint, "-n", team})
}
