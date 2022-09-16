package dependabot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRender(t *testing.T) {
	t.Run("Renders config", func(t *testing.T) {
		schedule := Schedule{"daily"}
		expectedConfig := Config{
			Version: 2,
			Updates: []Dependency{
				{"github-actions", "/", schedule},
				{"docker", "/", schedule},
				{"docker", "/a", schedule},
				{"npm", "/a/b/c/d", schedule},
			},
		}
		config := NewRender().Render(MatchedPaths{
			"Dockerfile":        "docker",
			"a/Dockerfile":      "docker",
			"a/b/c/d/yarn.lock": "npm",
			"/":                 "github-actions",
		})
		assert.Equal(t, expectedConfig, config)
	})
}
