package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	tasks "github.com/springernature/halfpipe/linters/triggers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type triggersLinter struct {
	fs              afero.Afero
	workingDir      string
	branchResolver  project.GitBranchResolver
	repoURIResolver project.RepoURIResolver
	gitLinter       func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error)
	cronLinter      func(cron manifest.TimerTrigger) (errs []error, warnings []error)
	dockerLinter    func(docker manifest.DockerTrigger) (errs []error, warnings []error)
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
		errs = append(errs, errors.NewTriggerError("git"))
	}

	if numCron > 1 {
		errs = append(errs, errors.NewTriggerError("cron"))
	}

	if numDocker > 1 {
		errs = append(errs, errors.NewTriggerError("docker"))
	}

	return
}

func (t triggersLinter) lintTrigger(man manifest.Manifest) (errs []error, warnings []error) {
	for i, trigger := range man.Triggers {

		prefixErrors := prefixErrorsWithIndex(fmt.Sprintf("triggers[%v]", i))

		var e, w []error
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			e, w = t.gitLinter(trigger, t.fs, t.workingDir, t.branchResolver, t.repoURIResolver)
		case manifest.TimerTrigger:
			e, w = t.cronLinter(trigger)
		case manifest.DockerTrigger:
			e, w = t.dockerLinter(trigger)
		}

		errs = append(errs, prefixErrors(e)...)
		warnings = append(warnings, prefixErrors(w)...)
	}
	return
}

func (t triggersLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Triggers Linter"
	result.DocsURL = "https://docs.halfpipe.io/manifest/triggers"

	result.Errors = append(result.Errors, t.lintOnlyOneOfEach(manifest.Triggers)...)
	if len(result.Errors) > 0 {
		return
	}

	errs, warnings := t.lintTrigger(manifest)
	result.Errors = append(result.Errors, errs...)
	result.Warnings = append(result.Warnings, warnings...)

	return
}

func NewTriggersLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) triggersLinter {
	return triggersLinter{
		fs:              fs,
		workingDir:      workingDir,
		branchResolver:  branchResolver,
		repoURIResolver: repoURIResolver,
		gitLinter:       tasks.LintGitTrigger,
		cronLinter:      tasks.LintCronTrigger,
		dockerLinter:    tasks.LintDockerTrigger,
	}
}
