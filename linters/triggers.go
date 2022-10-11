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
	gitLinter       func(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) []error
	cronLinter      func(cron manifest.TimerTrigger) []error
	dockerLinter    func(docker manifest.DockerTrigger) []error
	pipelineLinter  func(man manifest.Manifest, pipeline manifest.PipelineTrigger) []error
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

func (t triggersLinter) lintTrigger(man manifest.Manifest) (errs []error) {
	for i, trigger := range man.Triggers {

		wrapWithIndex := wrapErrorsWithIndex(fmt.Sprintf("triggers[%v]", i))

		var e []error
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			e = t.gitLinter(trigger, t.fs, t.workingDir, t.branchResolver, t.repoURIResolver, man.Platform)
		case manifest.TimerTrigger:
			e = t.cronLinter(trigger)
		case manifest.DockerTrigger:
			e = t.dockerLinter(trigger)
		case manifest.PipelineTrigger:
			e = t.pipelineLinter(man, trigger)
		}

		errs = append(errs, wrapWithIndex(e)...)
	}
	return errs
}

func (t triggersLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Triggers"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest#triggers"

	errs := t.lintOnlyOneOfEach(manifest.Triggers)
	if len(errs) > 0 {
		result.Add(errs...)
		return result
	}

	errs = t.lintTrigger(manifest)
	result.Add(errs...)
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
