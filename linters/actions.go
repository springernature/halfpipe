package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

type ActionsLinter struct{}

func NewActionsLinter() Linter {
	return ActionsLinter{}
}

func (linter ActionsLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "GitHub Actions"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/github-actions/overview/"
	if man.Platform.IsActions() {
		result.AddWarning(unsupportedTasks(man.Tasks, man, "tasks")...)
		result.AddWarning(unsupportedTriggers(man.Triggers)...)
		result.AddWarning(unsupportedFeatures(man.FeatureToggles)...)
	}
	return result
}

func unsupportedTasks(tasks manifest.TaskList, man manifest.Manifest, taskListId string) (errors []error) {
	for i, t := range tasks {
		taskIdx := fmt.Sprintf("%s[%v]", taskListId, i)

		appendError := func(err error) {
			errors = append(errors, fmt.Errorf("%s %w", taskIdx, err))
		}

		switch task := t.(type) {
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks, man, taskIdx)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks, man, taskIdx)...)
		default:
			if task.IsManualTrigger() {
				appendError(ErrUnsupportedManualTrigger)
			}
		}

		switch task := t.(type) {
		case manifest.DeployCF:
			if task.Rolling {
				appendError(ErrUnsupportedRolling)
			}
		case manifest.DockerPush:
			for _, trigger := range man.Triggers {
				if t, ok := trigger.(manifest.DockerTrigger); ok {
					if t.Image == task.Image {
						appendError(ErrDockerTriggerLoop.WithValue(t.Image))
					}
				}
			}
		case manifest.ConsumerIntegrationTest:
			if task.UseCovenant {
				appendError(ErrUnsupportedCovenant)
			}
		}

	}
	return errors
}

func unsupportedTriggers(triggers manifest.TriggerList) (errors []error) {
	for i, trigger := range triggers {
		appendError := func(err error) {
			errors = append(errors, fmt.Errorf("triggers[%v] %w", i, err))
		}

		switch t := trigger.(type) {
		case manifest.GitTrigger:
			if t.PrivateKey != "" {
				appendError(ErrUnsupportedGitPrivateKey)
			}
			if t.URI != "" {
				appendError(ErrUnsupportedGitUri)
			}
		case manifest.PipelineTrigger:
			appendError(ErrUnsupportedPipelineTrigger)
		default:
			// ok
		}
	}
	return errors
}

func unsupportedFeatures(features manifest.FeatureToggles) (errors []error) {
	if features.UpdatePipeline() {
		errors = []error{ErrUnsupportedUpdatePipeline}
	}
	return errors
}
