package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestOpsLevelSystem(t *testing.T) {
	t.Run("warning when opslevel.yml not found", func(t *testing.T) {
		man := manifest.Manifest{}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrOpsLevelNotFound)
	})

	t.Run("warning when opslevel.yml is invalid", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", ParseError: "failed to parse opslevel.yml: yaml: did not find expected ',' or ']'"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrOpsLevelInvalid)
	})

	t.Run("warning when system is empty", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("opslevel.system"))
	})

	t.Run("no warning when system matches pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", System: "appl-428"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.False(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
	})

	t.Run("warning when system does not match pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", System: "oscar"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("opslevel.system"))
	})
}
