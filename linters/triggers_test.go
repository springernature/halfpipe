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
	linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
		return
	}
	linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error, warnings []error) {
		return
	}
	linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error, warnings []error) {
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
		assertNotContainsError(t, result.Errors, ErrMultipleTriggers)
		assertNotContainsError(t, result.Warnings, ErrMultipleTriggers)
	})

	t.Run("with more than one of each there should be errors", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.DockerTrigger{},
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
				manifest.GitTrigger{},
				manifest.DockerTrigger{},
				manifest.TimerTrigger{},
			},
		}

		result := linter.Lint(man)
		assert.Len(t, result.Errors, 3)
		assertContainsError(t, result.Errors, ErrMultipleTriggers)
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
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error, warnings []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error, warnings []error) {
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
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error, warnings []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error, warnings []error) {
			numCallsCronTriggerLinter++
			return
		}
		linter.pipelineLinter = func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error, warnings []error) {
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
	gitError := errors.New("gitError")
	cronError := errors.New("cronError")
	cronWarning := errors.New("cronWarning")
	dockerError := errors.New("dockerError")
	dockerWarning := errors.New("dockerWarning")
	pipelineError := errors.New("pipelineError")
	pipelineWarning := errors.New("pipelineWarning")

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
			manifest.TimerTrigger{},
			manifest.DockerTrigger{},
			manifest.PipelineTrigger{},
		},
	}

	linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
	linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
		errs = append(errs, dockerError)
		warnings = append(warnings, dockerWarning)
		return
	}
	linter.gitLinter = func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error, warnings []error) {
		errs = append(errs, gitError)
		return
	}
	linter.cronLinter = func(cron manifest.TimerTrigger) (errs []error, warnings []error) {
		errs = append(errs, cronError)
		warnings = append(warnings, cronWarning)
		return
	}
	linter.pipelineLinter = func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error, warnings []error) {
		errs = append(errs, pipelineError)
		warnings = append(warnings, pipelineWarning)
		return
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assert.Len(t, result.Warnings, 3)

	assert.Equal(t, result.Errors[0].Error(), errors.New("triggers[0] gitError").Error())
	assert.Equal(t, result.Errors[1].Error(), errors.New("triggers[1] cronError").Error())
	assert.Equal(t, result.Errors[2].Error(), errors.New("triggers[2] dockerError").Error())
	assert.Equal(t, result.Errors[3].Error(), errors.New("triggers[3] pipelineError").Error())

	assert.Equal(t, result.Warnings[0].Error(), errors.New("triggers[1] cronWarning").Error())
	assert.Equal(t, result.Warnings[1].Error(), errors.New("triggers[2] dockerWarning").Error())
	assert.Equal(t, result.Warnings[2].Error(), errors.New("triggers[3] pipelineWarning").Error())
}
