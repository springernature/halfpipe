package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/linters/triggers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type triggersLinter struct {
	fs                         afero.Afero
	deprecatedDockerRegistries []string
	workingDir                 string
	branchResolver             project.GitBranchResolver
	repoURIResolver            project.RepoURIResolver
	gitLinter                  func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error)
	cronLinter                 func(cron manifest.TimerTrigger) (errs []error, warnings []error)
	dockerLinter               func(docker manifest.DockerTrigger, deprecatedDockerRegistries []string) (errs []error, warnings []error)
	pipelineLinter             func(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error, warnings []error)
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
		errs = append(errs, linterrors.NewTriggerError("git"))
	}

	if numCron > 1 {
		errs = append(errs, linterrors.NewTriggerError("cron"))
	}

	if numDocker > 1 {
		errs = append(errs, linterrors.NewTriggerError("docker"))
	}

	return errs
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
			e, w = t.dockerLinter(trigger, t.deprecatedDockerRegistries)
		case manifest.PipelineTrigger:
			e, w = t.pipelineLinter(man, trigger)
		}

		errs = append(errs, prefixErrors(e)...)
		warnings = append(warnings, prefixErrors(w)...)
	}
	return errs, warnings
}

func (t triggersLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Triggers"
	result.DocsURL = "https://docs.halfpipe.io/manifest/triggers"

	result.Errors = append(result.Errors, t.lintOnlyOneOfEach(manifest.Triggers)...)
	if len(result.Errors) > 0 {
		return result
	}

	errs, warnings := t.lintTrigger(manifest)
	result.Errors = append(result.Errors, errs...)
	result.Warnings = append(result.Warnings, warnings...)

	return result
}

func NewTriggersLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, deprecatedDockerRegistries []string) triggersLinter {
	return triggersLinter{
		fs:                         fs,
		deprecatedDockerRegistries: deprecatedDockerRegistries,
		workingDir:                 workingDir,
		branchResolver:             branchResolver,
		repoURIResolver:            repoURIResolver,
		gitLinter:                  triggers.LintGitTrigger,
		cronLinter:                 triggers.LintCronTrigger,
		dockerLinter:               triggers.LintDockerTrigger,
		pipelineLinter:             triggers.LintPipelineTrigger,
	}
}
