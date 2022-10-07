package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type triggersLinter struct {
	fs              afero.Afero
	workingDir      string
	branchResolver  project.GitBranchResolver
	repoURIResolver project.RepoURIResolver
	gitLinter       func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error, warnings []error)
	cronLinter      func(cron manifest.TimerTrigger) (errs []error, warnings []error)
	dockerLinter    func(docker manifest.DockerTrigger) (errs []error, warnings []error)
	pipelineLinter  func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error, warnings []error)
}

func (t triggersLinter) lintOnlyOneOfEach(triggers manifest.TriggerList) (errs []error) {
	numGit := 0
	numCron := 0
	numDocker := 0

	for _, trigger := range triggers {
		switch trigger.(type) {
		case manifest.GitTrigger:
			numGit++
		case manifest.TimerTrigger:
			numCron++
		case manifest.DockerTrigger:
			numDocker++
		}
	}

	if numGit > 1 {
		errs = append(errs, ErrMultipleTriggers.WithValue(manifest.GitTrigger{}.GetTriggerName()))
	}

	if numCron > 1 {
		errs = append(errs, ErrMultipleTriggers.WithValue(manifest.TimerTrigger{}.GetTriggerName()))
	}

	if numDocker > 1 {
		errs = append(errs, ErrMultipleTriggers.WithValue(manifest.DockerTrigger{}.GetTriggerName()))
	}

	return errs
}

func (t triggersLinter) lintTrigger(man manifest.Manifest) (errs []error, warnings []error) {
	for i, trigger := range man.Triggers {

		prefixErrors := prefixErrorsWithIndex(fmt.Sprintf("triggers[%v]", i))

		var e, w []error
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			e, w = t.gitLinter(trigger, t.fs, t.workingDir, t.branchResolver, t.repoURIResolver, man.Platform)
		case manifest.TimerTrigger:
			e, w = t.cronLinter(trigger)
		case manifest.DockerTrigger:
			e, w = t.dockerLinter(trigger)
		case manifest.PipelineTrigger:
			e, w = t.pipelineLinter(man, trigger)
		}

		errs = append(errs, prefixErrors(e)...)
		warnings = append(warnings, prefixErrors(w)...)
	}
	return errs, warnings
}

func (t triggersLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Triggers"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest#triggers"

	result.Errors = append(result.Errors, t.lintOnlyOneOfEach(manifest.Triggers)...)
	if len(result.Errors) > 0 {
		return result
	}

	errs, warnings := t.lintTrigger(manifest)
	result.Errors = append(result.Errors, errs...)
	result.Warnings = append(result.Warnings, warnings...)

	return result
}

func NewTriggersLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) triggersLinter {
	return triggersLinter{
		fs:              fs,
		workingDir:      workingDir,
		branchResolver:  branchResolver,
		repoURIResolver: repoURIResolver,
		gitLinter:       LintGitTrigger,
		cronLinter:      LintCronTrigger,
		dockerLinter:    LintDockerTrigger,
		pipelineLinter:  LintPipelineTrigger,
	}
}
