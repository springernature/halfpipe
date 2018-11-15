package linters

import (
	"path/filepath"
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
	repoUriResolver func() (string, error)
}

func NewRepoLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver) repoLinter {
	return repoLinter{
		Fs:              fs,
		WorkingDir:      workingDir,
		branchResolver:  branchResolver,
		repoUriResolver: gitconfig.OriginURL,
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

func (r repoLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Repo"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#repo"

	if man.Repo.URI == "" {
		result.AddError(errors.NewMissingField("repo.uri"))
		return
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, man.Repo.URI)
	if !match {
		result.AddError(errors.NewInvalidField("repo.uri", fmt.Sprintf("'%s' is not a valid git URI. If you are using SSH-aliases you must manually specify this field.", man.Repo.URI)))
		return
	}

	if strings.HasPrefix(man.Repo.URI, "git@") && man.Repo.PrivateKey == "" {
		result.AddError(errors.NewMissingField("repo.private_key"))
	}

	if strings.HasPrefix(man.Repo.URI, "http") && man.Repo.PrivateKey != "" {
		result.AddError(errors.NewInvalidField("repo.uri", "should be a ssh git url when private_key is set"))
	}

	if strings.HasPrefix(man.Repo.URI, "https") {
		result.AddWarning(fmt.Errorf("only public repos are supported with http(s). For private repos specify repo.uri with ssh"))
	}

	for _, glob := range append(man.Repo.WatchedPaths, man.Repo.IgnoredPaths...) {
		if err := r.checkGlob(glob, man.Repo.BasePath); err != nil {
			result.AddError(err)
		}
	}

	if man.Repo.GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(man.Repo.GitCryptKey) {
		result.AddError(errors.NewInvalidField("repo.git_crypt_key", "must be a vault secret"))
	}

	if currentBranch, err := r.branchResolver(); err != nil {
		result.AddError(err)
	} else {
		if currentBranch != "master" && man.Repo.Branch == "" {
			result.AddError(errors.NewInvalidField("repo.branch", "must be set if you are executing halfpipe from a non master branch"))
		}

		if man.Repo.Branch != currentBranch && man.Repo.Branch != "" {
			result.AddError(errors.NewInvalidField("repo.branch", fmt.Sprintf("You are currently on branch '%s' but you specified branch '%s'", currentBranch, man.Repo.Branch)))
		}
	}

	if resolvedRepoURI, err := r.repoUriResolver(); err != nil {
		result.AddError(err)
	} else {
		if resolvedRepoURI != man.Repo.URI {
			result.AddWarning(fmt.Errorf("you have specified 'repo.uri', make sure that its the same repo that you execute halfpipe in"))
		}
	}

	return
}
