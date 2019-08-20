package linters

import (
	"path/filepath"
	"reflect"
	"strings"

	"regexp"

	"fmt"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"github.com/tcnksm/go-gitconfig"
)

type repoLinter struct {
	Fs              afero.Afero
	WorkingDir      string
	branchResolver  project.GitBranchResolver
	repoURIResolver func() (string, error)
}

func NewRepoLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver) repoLinter {
	return repoLinter{
		Fs:              fs,
		WorkingDir:      workingDir,
		branchResolver:  branchResolver,
		repoURIResolver: gitconfig.OriginURL,
	}
}

func (r repoLinter) checkGlob(glob string, basePath string) error {
	repoRoot := strings.TrimSuffix(r.WorkingDir, basePath)

	matches, err := afero.Glob(r.Fs, filepath.Join(repoRoot, glob))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return errors.NewFileError(glob, "Could not find any files or directories matching glob")
	}
	return nil
}

func (r repoLinter) lintOnlyOneGitTrigger(man manifest.Manifest) error {
	numGitTriggers := 0
	var gitTrigger manifest.Trigger

	for _, trigger := range man.Triggers {
		switch trigger.(type) {
		case manifest.GitTrigger:
			gitTrigger = trigger
			numGitTriggers++
		}
	}

	if numGitTriggers > 1 {
		return errors.NewInvalidField("triggers", "You are only allowed one git trigger")
	}

	if !reflect.DeepEqual(man.Repo, manifest.Repo{}) && numGitTriggers != 0 && !reflect.DeepEqual(gitTrigger, manifest.GitTrigger{}) {
		return errors.NewInvalidField("repo/triggers", "You are only allowed to configure git with either repo or triggers")
	}

	return nil
}

func (r repoLinter) getValues(man manifest.Manifest) (URI, PrivateKey, Branch, BasePath, GitCryptKey, Prefix string, WatchedPaths, IgnoredPaths []string) {
	if !reflect.DeepEqual(man.Repo, manifest.Repo{}) {
		return man.Repo.URI, man.Repo.PrivateKey, man.Repo.Branch, man.Repo.BasePath, man.Repo.GitCryptKey, "repo", man.Repo.WatchedPaths, man.Repo.IgnoredPaths
	}

	// The first thing we do in the linter is to make sure that we use
	// either manifest.repo or manifest.triggers for git and that there
	// is only one git trigger, thus we can assume that the first git
	// trigger we find will be the correct one
	for index, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			return trigger.URI, trigger.PrivateKey, trigger.Branch, trigger.BasePath, trigger.GitCryptKey, fmt.Sprintf("triggers[%d]", index), trigger.WatchedPaths, trigger.IgnoredPaths
		}
	}
	return
}

func (r repoLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Repo"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#repo"

	if err := r.lintOnlyOneGitTrigger(man); err != nil {
		result.AddError(err)
		return
	}

	URI, PrivateKey, Branch, BasePath, GitCryptKey, Prefix, WatchedPaths, IgnoredPaths := r.getValues(man)

	if URI == "" {
		result.AddError(errors.NewMissingField("repo.uri/triggers[x].uri"))
		return
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, URI)
	if !match {
		result.AddError(errors.NewInvalidField(fmt.Sprintf("%s.uri", Prefix), fmt.Sprintf("'%s' is not a valid git URI. If you are using SSH-aliases you must manually specify this field.", URI)))
		return
	}

	if strings.HasPrefix(URI, "git@") && PrivateKey == "" {
		result.AddError(errors.NewMissingField(fmt.Sprintf("%s.private_key", Prefix)))
	}

	if strings.HasPrefix(URI, "http") && PrivateKey != "" {
		result.AddError(errors.NewInvalidField(fmt.Sprintf("%s.uri", Prefix), "should be a ssh git url when private_key is set"))
	}

	if strings.HasPrefix(URI, "https") {
		result.AddWarning(fmt.Errorf("only public repos are supported with http(s). For private repos specify %s.uri with ssh", Prefix))
	}

	for _, glob := range append(WatchedPaths, IgnoredPaths...) {

		if err := r.checkGlob(glob, BasePath); err != nil {
			result.AddError(err)
		}
	}

	if GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(GitCryptKey) {
		result.AddError(errors.NewInvalidField(fmt.Sprintf("%s.git_crypt_key", Prefix), "must be a vault secret"))
	}

	if currentBranch, err := r.branchResolver(); err != nil {
		result.AddError(err)
	} else {

		if currentBranch != "master" && Branch == "" {
			result.AddError(errors.NewInvalidField(fmt.Sprintf("%s.branch", Prefix), "must be set if you are executing halfpipe from a non master branch"))
		}

		if Branch != currentBranch && Branch != "" {
			result.AddError(errors.NewInvalidField(fmt.Sprintf("%s.branch", Prefix), fmt.Sprintf("You are currently on branch '%s' but you specified branch '%s'", currentBranch, Branch)))
		}
	}

	if resolvedRepoURI, err := r.repoURIResolver(); err != nil {
		result.AddError(err)
	} else {
		if resolvedRepoURI != URI {
			result.AddWarning(fmt.Errorf("you have specified '%s.uri', make sure that its the same repo that you execute halfpipe in", Prefix))
		}
	}

	return
}
