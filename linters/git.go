package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"path/filepath"
	"regexp"
	"strings"
)

func checkGlob(glob string, basePath, workingDir string, fs afero.Afero) error {
	// todo: fix stupid hack because
	// basePath always has /
	// workingDir has win or unix slashes
	unixWorkingDir := strings.ReplaceAll(workingDir, `\`, `/`)
	repoRoot := strings.TrimSuffix(unixWorkingDir, basePath)

	matches, err := afero.Glob(fs, filepath.Join(repoRoot, glob))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return ErrFileNotFound.WithValue(glob)
	}
	return nil
}

func LintGitTrigger(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver, platform manifest.Platform) (errs []error) {
	if platform.IsConcourse() {
		if git.URI == "" {
			errs = append(errs, NewErrMissingField("uri"))
			return errs
		}

		match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, git.URI)
		if !match {
			errs = append(errs, NewErrInvalidField("uri", fmt.Sprintf("'%s' is not a valid git URI. If you are using SSH-aliases you must manually specify this field.", git.URI)))
			return errs
		}

		if strings.HasPrefix(git.URI, "git@") && git.PrivateKey == "" {
			errs = append(errs, NewErrMissingField("private_key"))
		}

		if strings.HasPrefix(git.URI, "http") && git.PrivateKey != "" {
			errs = append(errs, NewErrInvalidField("uri", "should be a ssh git url when private_key is set"))
		}

		if strings.HasPrefix(git.URI, "https") {
			errs = append(errs, NewErrInvalidField("uri", "only public repos are supported with http(s). For private repos specify uri with ssh").AsWarning())
		}

		if git.GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(git.GitCryptKey) {
			errs = append(errs, NewErrInvalidField("git_crypt_key", "must be a vault secret"))
		}
	}

	for _, glob := range git.WatchedPaths {
		if err := checkGlob(glob, git.BasePath, workingDir, fs); err != nil {
			errs = append(errs, err)
		}
	}

	if currentBranch, err := branchResolver(); err != nil {
		errs = append(errs, NewErrExternal(err).AsWarning())
	} else {

		if config.CheckBranch == "true" && platform.IsConcourse() {
			if git.Branch == "" {
				// We default to either `main`, `master` or nothing in defaultGitTrigger
				errs = append(errs, NewErrInvalidField("branch", "must be set if you are executing halfpipe from a non main/master branch"))
			}

			if git.Branch != currentBranch && git.Branch != "" {
				errs = append(errs, NewErrInvalidField("branch", fmt.Sprintf("you are currently on branch '%s' but you specified branch '%s'", currentBranch, git.Branch)))
			}
		}
	}

	if resolvedRepoURI, err := repoURIResolver(); err != nil {
		errs = append(errs, NewErrExternal(err).AsWarning())
	} else {
		if resolvedRepoURI != git.URI && platform.IsConcourse() {
			errs = append(errs, NewErrInvalidField("uri", "must be uri of the repo that you execute halfpipe in").AsWarning())
		}
	}

	return errs
}
