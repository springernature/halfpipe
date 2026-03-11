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

		opsLevel, found, err := ParseOpsLevel(fs, "/repo", "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "my-app", opsLevel.Name)
		assert.Equal(t, "appl-123", opsLevel.System)
	})

	t.Run("returns no error when file does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		_, found, err := ParseOpsLevel(fs, "/repo", "/repo")
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`not: [valid: yaml`), 0644)

		_, found, err := ParseOpsLevel(fs, "/repo", "/repo")
		assert.Error(t, err)
		assert.True(t, found)
	})

	t.Run("walks up to git root to find opslevel.yml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  name: root-app
  system: appl-root
`), 0644)

		opsLevel, found, err := ParseOpsLevel(fs, "/repo/a/b/c", "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
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

		opsLevel, found, err := ParseOpsLevel(fs, "/repo/sub", "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "sub-app", opsLevel.Name)
	})

	t.Run("does not walk past gitRootDir", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/opslevel.yml", []byte(`
version: 1
component:
  name: above-root
`), 0644)

		_, found, err := ParseOpsLevel(fs, "/repo/sub", "/repo")
		assert.NoError(t, err)
		assert.False(t, found)
	})
}
