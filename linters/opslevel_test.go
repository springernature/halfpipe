package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestOpsLevelSystem(t *testing.T) {
	t.Run("no warning when system is empty", func(t *testing.T) {
		man := manifest.Manifest{}
		result := opsLevelLinter{}.Lint(man)
		assert.False(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
	})

	t.Run("no warning when system matches pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{System: "appl-428"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.False(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
	})

	t.Run("no warning for appl-0", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{System: "appl-0"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.False(t, result.HasWarnings())
	})

	t.Run("warning when system does not match pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{System: "oscar"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("opslevel.system"))
	})

	t.Run("warning when system is missing appl- prefix", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{System: "428"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("opslevel.system"))
	})

	t.Run("warning when system has extra characters", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{System: "appl-428-extra"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasWarnings())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("opslevel.system"))
	})
}
