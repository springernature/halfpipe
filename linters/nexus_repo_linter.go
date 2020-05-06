package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"os"
	"path/filepath"
	"strings"
)

type nexusRepoLinter struct {
	fs afero.Afero
}

func (l nexusRepoLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Deprecated Nexus Repository"
	result.DocsURL = "https://ee-discourse.springernature.io/t/repo-tools-nexus-repository-end-of-life/1318"

	repoTools := "repo.tools.springer-sbm.com"

	filenamesToCheck := []string{
		"build.gradle",
		"build.sbt",
		"Build.scala",
		"plugins.sbt",
		"Dependencies.scala",
		"pom.xml",
	}

	shouldCheckFile := func(name string) bool {
		for _, filename := range filenamesToCheck {
			if strings.ToLower(filename) == strings.ToLower(name) {
				return true
			}
		}
		return false
	}

	_ = l.fs.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if shouldCheckFile(info.Name()) {
			if isMatch, _ := l.fs.FileContainsBytes(path, []byte(repoTools)); isMatch {
				errOrWarning := fmt.Errorf("file '%s' references '%s'", path, repoTools)
				if man.FeatureToggles.DisableDeprecatedNexusRepositoryError() {
					result.AddWarning(errOrWarning)
				} else {
					result.AddError(fmt.Errorf("%s. To supress this error use the feature toggle as described in <%s>, you have until 24 August 2020 to migrate", errOrWarning.Error(), result.DocsURL))
				}
			}
		}
		return nil
	})

	return
}

func NewNexusRepoLinter(fs afero.Afero) Linter {
	return nexusRepoLinter{
		fs: fs,
	}
}
