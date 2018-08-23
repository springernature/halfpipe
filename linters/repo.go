package linters

import (
	"path/filepath"
	"strings"

	"regexp"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type repoLinter struct {
	Fs             afero.Afero
	WorkingDir     string
	BranchResolver project.GitBranchResolver
}

func NewRepoLinter(fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver) repoLinter {
	return repoLinter{fs, workingDir, branchResolver}
}

func (r repoLinter) checkGlob(glob string, basePath string) error {

	//need the path to the repo
	repoRoot := strings.Replace(r.WorkingDir, basePath, "", -1)

	matches, err := afero.Glob(r.Fs, filepath.Join(repoRoot, glob))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return errors.NewFileError(glob, "Could not find any files or directories matching glob")
	}
	return nil
}

func (r repoLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Repo"
	result.DocsURL = "https://docs.halfpipe.io/docs/manifest/#repo"

	if man.Repo.URI == "" {
		result.AddError(errors.NewMissingField("repo.uri"))
		return
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, man.Repo.URI)
	if !match {
		result.AddError(errors.NewInvalidField("repo.uri", "must be a valid git uri"))
		return
	}

	if strings.HasPrefix(man.Repo.URI, "git@") && man.Repo.PrivateKey == "" {
		result.AddError(errors.NewMissingField("repo.private_key"))
	}

	if strings.HasPrefix(man.Repo.URI, "http") && man.Repo.PrivateKey != "" {
		result.AddError(errors.NewInvalidField("repo.uri", "should be a ssh git url when private_key is set"))
	}

	for _, glob := range append(man.Repo.WatchedPaths, man.Repo.IgnoredPaths...) {
		if err := r.checkGlob(glob, man.Repo.BasePath); err != nil {
			result.AddError(err)
		}
	}

	if man.Repo.GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(man.Repo.GitCryptKey) {
		result.AddError(errors.NewInvalidField("repo.git_crypt_key", "must be a vault secret"))
	}

	return
}
