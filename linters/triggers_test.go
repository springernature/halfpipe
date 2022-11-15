package linters

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLintOnlyOneOfEachAllowed(t *testing.T) {
	linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
	linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error) {
		return
	}
	linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error) {
		return
	}
	linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error) {
		return
	}

	t.Run("with only one of each there should be no errors", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
				manifest.DockerTrigger{},
			},
		}

		result := linter.Lint(man)
		assertNotContainsError(t, result.Issues, ErrMultipleTriggers)
	})

	t.Run("multiple pipeline and docker triggers is ok", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.DockerTrigger{},
				manifest.PipelineTrigger{},
				manifest.DockerTrigger{},
				manifest.PipelineTrigger{},
			},
		}

		result := linter.Lint(man)
		assertNotContainsError(t, result.Issues, ErrMultipleTriggers)
	})

	t.Run("with more than one of each there should be errors", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
			},
		}

		result := linter.Lint(man)
		assert.Len(t, result.Issues, 2)
		assertContainsError(t, result.Issues, ErrMultipleTriggers)
	})
}

func TestCallsOutCorrectly(t *testing.T) {
	t.Run("no triggers", func(t *testing.T) {
		numCallsGitTriggerLinter := 0
		numCallsCronTriggerLinter := 0
		numCallsDockerTriggerLinter := 0
		numCallsPipelineTriggerLinter := 0

		man := manifest.Manifest{}

		linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error) {
			numCallsCronTriggerLinter++
			return
		}

		linter.Lint(man)
		assert.Equal(t, 0, numCallsCronTriggerLinter)
		assert.Equal(t, 0, numCallsGitTriggerLinter)
		assert.Equal(t, 0, numCallsDockerTriggerLinter)
		assert.Equal(t, 0, numCallsPipelineTriggerLinter)
	})

	t.Run("all triggers", func(t *testing.T) {

		numCallsGitTriggerLinter := 0
		numCallsCronTriggerLinter := 0
		numCallsDockerTriggerLinter := 0
		numCallsPipelineTriggerLinter := 0

		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
				manifest.PipelineTrigger{},
				manifest.TimerTrigger{},
				manifest.DockerTrigger{},
				manifest.PipelineTrigger{},
			},
		}

		linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error) {
			numCallsCronTriggerLinter++
			return
		}
		linter.pipelineLinter = func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error) {
			numCallsPipelineTriggerLinter++
			return
		}

		linter.Lint(man)
		assert.Equal(t, 1, numCallsCronTriggerLinter)
		assert.Equal(t, 1, numCallsGitTriggerLinter)
		assert.Equal(t, 1, numCallsDockerTriggerLinter)
		assert.Equal(t, 2, numCallsPipelineTriggerLinter)
	})

}

func TestReturnsErrorsCorrectlyAndWithIndexedPrefix(t *testing.T) {
	gitError := newError("gitError")
	cronError := newError("cronError")
	dockerError := newError("dockerError")
	pipelineError := newError("pipelineError")

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
			manifest.TimerTrigger{},
			manifest.DockerTrigger{},
			manifest.PipelineTrigger{},
		},
	}

	linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
	linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error) {
		errs = append(errs, dockerError)
		return
	}
	linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error) {
		errs = append(errs, gitError)
		return
	}
	linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error) {
		errs = append(errs, cronError)
		return
	}
	linter.pipelineLinter = func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error) {
		errs = append(errs, pipelineError)
		return
	}

	result := linter.Lint(man)
	assert.Len(t, result.Issues, 4)
	assert.Equal(t, result.Issues[0].Error(), errors.New("triggers[0] gitError").Error())
	assert.Equal(t, result.Issues[1].Error(), errors.New("triggers[1] cronError").Error())
	assert.Equal(t, result.Issues[2].Error(), errors.New("triggers[2] dockerError").Error())
	assert.Equal(t, result.Issues[3].Error(), errors.New("triggers[3] pipelineError").Error())
}
