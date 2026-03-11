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
  name: oscar-bandiera
  description: Component for managing feature flags
  type: application_module
  owner: oscar
  language: kotlin
  framework: http4k
  lifecycle: core
  tier: mission-critical
  system: appl-428
`), 0644)

		opsLevel, found, err := ParseOpsLevel(fs, "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "oscar-bandiera", opsLevel.Name)
		assert.Equal(t, "appl-428", opsLevel.System)
	})

	t.Run("returns no error when file does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		_, found, err := ParseOpsLevel(fs, "/repo")
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`not: [valid: yaml`), 0644)

		_, found, err := ParseOpsLevel(fs, "/repo")
		assert.Error(t, err)
		assert.True(t, found)
		assert.Contains(t, err.Error(), "failed to parse opslevel.yml")
	})

	t.Run("returns empty OpsLevel when component fields are missing", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  description: just a description
`), 0644)

		opsLevel, found, err := ParseOpsLevel(fs, "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "", opsLevel.Name)
		assert.Equal(t, "", opsLevel.System)
	})

	t.Run("parses file with only name and system", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_ = fs.WriteFile("/repo/opslevel.yml", []byte(`
version: 1
component:
  name: my-app
  system: appl-123
`), 0644)

		opsLevel, found, err := ParseOpsLevel(fs, "/repo")
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "my-app", opsLevel.Name)
		assert.Equal(t, "appl-123", opsLevel.System)
	})
}
