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
	linter.gitLinter = func(man manifest.Manifest, git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error) {
		return
	}
	linter.cronLinter = func(man manifest.Manifest, cron manifest.TimerTrigger) (errs []error, warnings []error) {
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
		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Warnings, 0)
	})

	t.Run("with more than one of each there should be no errors", func(t *testing.T) {
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
		assert.Len(t, result.Warnings, 0)
		assertTriggerErrorInErrors(t, "git", result.Errors)
		assertTriggerErrorInErrors(t, "cron", result.Errors)
		assertTriggerErrorInErrors(t, "docker", result.Errors)
	})
}

func TestCallsOutCorrectly(t *testing.T) {
	t.Run("no triggers", func(t *testing.T) {
		numCallsGitTriggerLinter := 0
		numCallsCronTriggerLinter := 0
		numCallsDockerTriggerLinter := 0

		man := manifest.Manifest{}

		linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(man manifest.Manifest, git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(man manifest.Manifest, cron manifest.TimerTrigger) (errs []error, warnings []error) {
			numCallsCronTriggerLinter++
			return
		}

		linter.Lint(man)
		assert.Equal(t, 0, numCallsCronTriggerLinter)
		assert.Equal(t, 0, numCallsGitTriggerLinter)
		assert.Equal(t, 0, numCallsDockerTriggerLinter)
	})

	t.Run("all triggers", func(t *testing.T) {

		numCallsGitTriggerLinter := 0
		numCallsCronTriggerLinter := 0
		numCallsDockerTriggerLinter := 0

		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
				manifest.DockerTrigger{},
			},
		}

		linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
		linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
			numCallsDockerTriggerLinter++
			return
		}
		linter.gitLinter = func(man manifest.Manifest, git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error) {
			numCallsGitTriggerLinter++
			return
		}
		linter.cronLinter = func(man manifest.Manifest, cron manifest.TimerTrigger) (errs []error, warnings []error) {
			numCallsCronTriggerLinter++
			return
		}

		linter.Lint(man)
		assert.Equal(t, 1, numCallsCronTriggerLinter)
		assert.Equal(t, 1, numCallsGitTriggerLinter)
		assert.Equal(t, 1, numCallsDockerTriggerLinter)
	})

}

func TestReturnsErrorsCorrectlyAndWithIndexedPrefix(t *testing.T) {
	gitError := errors.New("gitError")
	cronError := errors.New("cronError")
	cronWarning := errors.New("cronWarning")
	dockerError := errors.New("dockerError")
	dockerWarning := errors.New("dockerWarning")

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
			manifest.TimerTrigger{},
			manifest.DockerTrigger{},
		},
	}

	linter := NewTriggersLinter(afero.Afero{}, "", nil, nil)
	linter.dockerLinter = func(docker manifest.DockerTrigger) (errs []error, warnings []error) {
		errs = append(errs, dockerError)
		warnings = append(warnings, dockerWarning)
		return
	}
	linter.gitLinter = func(man manifest.Manifest, git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error) {
		errs = append(errs, gitError)
		return
	}
	linter.cronLinter = func(man manifest.Manifest, cron manifest.TimerTrigger) (errs []error, warnings []error) {
		errs = append(errs, cronError)
		warnings = append(warnings, cronWarning)
		return
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 3)
	assert.Len(t, result.Warnings, 2)

	assert.Equal(t, result.Errors[0].Error(), errors.New("triggers[0] gitError").Error())
	assert.Equal(t, result.Errors[1].Error(), errors.New("triggers[1] cronError").Error())
	assert.Equal(t, result.Errors[2].Error(), errors.New("triggers[2] dockerError").Error())

	assert.Equal(t, result.Warnings[0].Error(), errors.New("triggers[1] cronWarning").Error())
	assert.Equal(t, result.Warnings[1].Error(), errors.New("triggers[2] dockerWarning").Error())
}