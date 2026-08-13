package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestOpsLevelSystem(t *testing.T) {
	t.Run("error when opslevel.yml not found", func(t *testing.T) {
		man := manifest.Manifest{}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrOpsLevelNotFound)
	})

	t.Run("error when opslevel.yml is invalid", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", ParseError: "failed to parse opslevel.yml: yaml: did not find expected ',' or ']'"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrOpsLevelInvalid)
	})

	t.Run("error when system is empty", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("component.system"))
	})

	t.Run("no error when system matches pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", System: "APPL-428"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.False(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
	})

	t.Run("error when system does not match pattern", func(t *testing.T) {
		man := manifest.Manifest{
			OpsLevel: manifest.OpsLevel{RelativePath: "opslevel.yml", System: "oscar"},
		}
		result := opsLevelLinter{}.Lint(man)
		assert.True(t, result.HasErrors())
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("component.system"))
	})
}
