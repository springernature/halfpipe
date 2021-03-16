package upload

import (
	"fmt"
	"strings"
	"testing"

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
triggers:
- type: git
  branch: %s
`, team, pipeline, branch)

var homedir = "/home/my-user"

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

func planToString(plan Plan) string {
	var parts []string
	for _, c := range plan {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, "\n")
}

func TestErrors(t *testing.T) {
	t.Run("when halfpipe file doesnt exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		planner := NewPlanner(fs, pathResolver, homedir, NullpipelineFile, false, "master", osResolver, envResolver, "", "")
		_, err := planner.Plan()

		assert.Error(t, err)
	})

	t.Run("when fly is not on path", func(t *testing.T) {
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
		}, false, "master", osResolver, envResolver, "", "")

		_, err := planner.Plan()
		assert.Equal(t, ErrFlyNotInstalled(osResolver()), err)
	})

	t.Run("when manifest doesnt contain team or pipeline", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile(".halfpipe.io", []byte(""), 0777)

		planner := NewPlanner(fs, pathResolver, homedir, NullpipelineFile, false, "master", osResolver, envResolver, "", "")
		_, err := planner.Plan()

		assert.Error(t, err)
	})
}

func TestOnMaster(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
	pipelineFile := func(fs afero.Afero) (afero.File, error) {
		file, _ := fs.Create("pipeline.yml")
		return file, nil
	}

	t.Run("returns a plan", func(t *testing.T) {
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, false, "master", osResolver, envResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := `halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c https://concourse.halfpipe.io -n my-team
fly -t my-team set-pipeline -p my-pipeline -c pipeline.yml --check-creds`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})

	t.Run("returns a plan when on main", func(t *testing.T) {
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, false, "main", osResolver, envResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := `halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c https://concourse.halfpipe.io -n my-team
fly -t my-team set-pipeline -p my-pipeline -c pipeline.yml --check-creds`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})

	t.Run("returns a non-interactive plan", func(t *testing.T) {
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, true, "master", osResolver, envResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := `halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c https://concourse.halfpipe.io -n my-team
fly -t my-team set-pipeline -p my-pipeline -c pipeline.yml --check-creds --non-interactive`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})

	t.Run("returns a plan with overridden concourse api", func(t *testing.T) {
		concourseEndpoint := "https://someRandom.location.com"
		overriddenEnvResolver := func(envVar string) string {
			return concourseEndpoint
		}

		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, true, "master", osResolver, overriddenEnvResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := fmt.Sprintf(`halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c %s -n my-team
fly -t my-team set-pipeline -p my-pipeline -c pipeline.yml --check-creds --non-interactive`, concourseEndpoint)

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})

}

func TestOnBranch(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(validPipelineWithBranch), 0777)
	pipelineFile := func(fs afero.Afero) (afero.File, error) {
		file, _ := fs.Create("pipeline.yml")
		return file, nil
	}

	t.Run("returns a plan with security question", func(t *testing.T) {
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, false, branch, osResolver, envResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := `# Security question
halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c https://concourse.halfpipe.io -n my-team
fly -t my-team set-pipeline -p my-pipeline-my-branch -c pipeline.yml --check-creds`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})

	t.Run("returns a non interactive plan", func(t *testing.T) {
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, true, branch, osResolver, envResolver, "", "")
		plan, err := planner.Plan()

		expectedPlan := `halfpipe > pipeline.yml
fly -t my-team status || fly -t my-team login -c https://concourse.halfpipe.io -n my-team
fly -t my-team set-pipeline -p my-pipeline-my-branch -c pipeline.yml --check-creds --non-interactive`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})
}

func TestUnpause(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pipelineFile := func(fs afero.Afero) (afero.File, error) {
		file, _ := fs.Create("pipeline.yml")
		return file, nil
	}

	t.Run("returns a plan on master", func(t *testing.T) {
		fs.WriteFile(".halfpipe.io", []byte(validPipeline), 0777)
		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, true, "master", osResolver, envResolver, "", "")
		plan, err := planner.Unpause()

		expectedPlan := `fly -t my-team unpause-pipeline -p my-pipeline`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))

	})

	t.Run("returns a plan on a branch", func(t *testing.T) {
		fs.WriteFile(".halfpipe.io", []byte(validPipelineWithBranch), 0777)

		planner := NewPlanner(fs, pathResolver, homedir, pipelineFile, true, branch, osResolver, envResolver, "", "")
		plan, err := planner.Unpause()

		expectedPlan := `fly -t my-team unpause-pipeline -p my-pipeline-my-branch`

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, planToString(plan))
	})
}
