package triggers

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"path/filepath"
	"regexp"
	"strings"
)

func checkGlob(glob string, basePath, workingDir string, fs afero.Afero) error {
	repoRoot := strings.TrimSuffix(workingDir, basePath)

	matches, err := afero.Glob(fs, filepath.Join(repoRoot, glob))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return linterrors.NewFileError(glob, "Could not find any files or directories matching glob")
	}
	return nil
}

func LintGitTrigger(git manifest.GitTrigger, fs afero.Afero, workingDir string, branchResolver project.GitBranchResolver, repoURIResolver project.RepoURIResolver) (errs []error, warnings []error) {
	if git.URI == "" {
		errs = append(errs, linterrors.NewMissingField("uri"))
		return
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, git.URI)
	if !match {
		errs = append(errs, linterrors.NewInvalidField("uri", fmt.Sprintf("'%s' is not a valid git URI. If you are using SSH-aliases you must manually specify this field.", git.URI)))
		return
	}

	if strings.HasPrefix(git.URI, "git@") && git.PrivateKey == "" {
		errs = append(errs, linterrors.NewMissingField("private_key"))
	}

	if strings.HasPrefix(git.URI, "http") && git.PrivateKey != "" {
		errs = append(errs, linterrors.NewInvalidField("uri", "should be a ssh git url when private_key is set"))
	}

	if strings.HasPrefix(git.URI, "https") {
		warnings = append(warnings, fmt.Errorf("only public repos are supported with http(s). For private repos specify uri with ssh"))
	}

	if git.GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(git.GitCryptKey) {
		errs = append(errs, linterrors.NewInvalidField("git_crypt_key", "must be a vault secret"))
	}

	for _, glob := range append(git.WatchedPaths, git.IgnoredPaths...) {
		if err := checkGlob(glob, git.BasePath, workingDir, fs); err != nil {
			errs = append(errs, err)
		}
	}

	if currentBranch, err := branchResolver(); err != nil {
		errs = append(errs, err)
	} else {

		if config.CheckBranch == "true" {
			if currentBranch != "master" && git.Branch == "" {
				errs = append(errs, linterrors.NewInvalidField("branch", "must be set if you are executing halfpipe from a non master branch"))
			}

			if git.Branch != currentBranch && git.Branch != "" {
				errs = append(errs, linterrors.NewInvalidField("branch", fmt.Sprintf("You are currently on branch '%s' but you specified branch '%s'", currentBranch, git.Branch)))
			}
		}
	}

	if resolvedRepoURI, err := repoURIResolver(); err != nil {
		errs = append(errs, err)
	} else {
		if resolvedRepoURI != git.URI {
			warnings = append(warnings, fmt.Errorf("you have specified 'uri', make sure that its the same repo that you execute halfpipe in"))
		}
	}

	return
}
