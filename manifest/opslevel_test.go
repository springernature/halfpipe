package manifest

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestParseOpsLevel(t *testing.T) {
	t.Run("parses valid opslevel.yml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  name: my-app
  system: appl-123
`), 0644)

		opsLevel := ParseOpsLevel(fs, "/repo", "/repo")
		assert.Equal(t, "opslevel.yml", opsLevel.RelativePath)
		assert.Empty(t, opsLevel.ParseError)
		assert.Equal(t, "my-app", opsLevel.Name)
		assert.Equal(t, "appl-123", opsLevel.System)
	})

	t.Run("not found when file does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		opsLevel := ParseOpsLevel(fs, "/repo", "/repo")
		assert.Empty(t, opsLevel.RelativePath)
		assert.Empty(t, opsLevel.ParseError)
	})

	t.Run("parse error for invalid yaml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`not: [valid: yaml`), 0644)

		opsLevel := ParseOpsLevel(fs, "/repo", "/repo")
		assert.Equal(t, "opslevel.yml", opsLevel.RelativePath)
		assert.NotEmpty(t, opsLevel.ParseError)
	})

	t.Run("walks up to git root to find opslevel.yml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  name: root-app
  system: appl-root
`), 0644)

		opsLevel := ParseOpsLevel(fs, "/repo/a/b/c", "/repo")
		assert.Equal(t, "../../../opslevel.yml", opsLevel.RelativePath)
		assert.Equal(t, "root-app", opsLevel.Name)
	})

	t.Run("closest opslevel.yml wins", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  name: root-app
`), 0644)
		_ = fs.WriteFile("/repo/sub/opslevel.yml", []byte(`
version: 1
component:
  name: sub-app
`), 0644)

		opsLevel := ParseOpsLevel(fs, "/repo/sub", "/repo")
		assert.Equal(t, "opslevel.yml", opsLevel.RelativePath)
		assert.Equal(t, "sub-app", opsLevel.Name)
	})

	t.Run("does not walk past gitRootDir", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/opslevel.yml", []byte(`
version: 1
component:
  name: above-root
`), 0644)

		opsLevel := ParseOpsLevel(fs, "/repo/sub", "/repo")
		assert.Empty(t, opsLevel.RelativePath)
	})
}
