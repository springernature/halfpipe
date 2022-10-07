package linters

import (
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
		result.AddWarning(unsupportedTasks(man.Tasks, man)...)
		result.AddWarning(unsupportedTriggers(man.Triggers)...)
		result.AddWarning(unsupportedFeatures(man.FeatureToggles)...)
	}
	return result
}

func unsupportedTasks(tasks manifest.TaskList, man manifest.Manifest) (errors []error) {
	for _, t := range tasks {
		switch task := t.(type) {
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
		default:
			if task.IsManualTrigger() {
				errors = append(errors, ErrUnsupportedManualTrigger)
			}
		}

		switch task := t.(type) {
		case manifest.DeployCF:
			if task.Rolling {
				errors = append(errors, ErrUnsupportedRolling)
			}
		case manifest.DockerPush:
			for _, trigger := range man.Triggers {
				if t, ok := trigger.(manifest.DockerTrigger); ok {
					if t.Image == task.Image {
						errors = append(errors, ErrDockerTriggerLoop.WithValue(t.Image))
					}
				}
			}
		case manifest.ConsumerIntegrationTest:
			if task.UseCovenant {
				errors = append(errors, ErrUnsupportedCovenant)
			}
		}

	}
	return errors
}

func unsupportedTriggers(triggers manifest.TriggerList) (errors []error) {
	for _, trigger := range triggers {
		switch t := trigger.(type) {
		case manifest.GitTrigger:
			if t.PrivateKey != "" {
				errors = append(errors, ErrUnsupportedGitPrivateKey)
			}
			if t.URI != "" {
				errors = append(errors, ErrUnsupportedGitUri)
			}
		case manifest.PipelineTrigger:
			errors = append(errors, ErrUnsupportedPipelineTrigger)
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
