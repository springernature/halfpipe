package linters

import (
	"github.com/spf13/afero"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestNexusRepoLinter(t *testing.T) {
	writeFile := func(fs afero.Afero, path string) {
		fs.WriteFile(path, []byte("this content should match because it contains repo.tools.springer-sbm.com."), 0777)
	}

	t.Run("no matching files", func(t *testing.T) {
		fs := afero.Afero{afero.NewMemMapFs()}
		writeFile(fs, "ignored.txt")

		linter := NewNexusRepoLinter(fs)
		result := linter.Lint(manifest.Manifest{})
		assert.Len(t, result.Errors, 0)
	})

	t.Run("one matching file", func(t *testing.T) {
		fs := afero.Afero{afero.NewMemMapFs()}
		writeFile(fs, "foo/bar/build.sbt")

		linter := NewNexusRepoLinter(fs)
		result := linter.Lint(manifest.Manifest{})
		assert.Len(t, result.Errors, 1)
	})

	t.Run("multiple matching files", func(t *testing.T) {
		fs := afero.Afero{afero.NewMemMapFs()}
		writeFile(fs, "foo/bar/build.sbt")
		writeFile(fs, "foo/bar/baz/pom.xml")

		linter := NewNexusRepoLinter(fs)
		result := linter.Lint(manifest.Manifest{})
		assert.Len(t, result.Errors, 2)
	})

	t.Run("feature toggle has no effect", func(t *testing.T) {
		fs := afero.Afero{afero.NewMemMapFs()}
		writeFile(fs, "foo/bar/build.sbt")
		writeFile(fs, "foo/bar/baz/pom.xml")

		linter := NewNexusRepoLinter(fs)
		result := linter.Lint(manifest.Manifest{FeatureToggles: []string{manifest.FeatureDisableDeprecatedNexusRepositoryError}})
		assert.Len(t, result.Errors, 2)
	})

	t.Run("ignore dot directories", func(t *testing.T) {
		fs := afero.Afero{afero.NewMemMapFs()}
		writeFile(fs, ".git/build.sbt")
		writeFile(fs, ".any-dot-dir/pom.xml")

		linter := NewNexusRepoLinter(fs)
		result := linter.Lint(manifest.Manifest{})
		assert.Len(t, result.Warnings, 0)
		assert.Len(t, result.Errors, 0)
	})
}
