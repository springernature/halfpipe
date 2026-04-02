package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Run("Renders config", func(t *testing.T) {
		schedule := Schedule{"daily"}
		cooldown := Cooldown{DefaultDays: 5}
		groups := Groups{
			"minor-and-patch": Group{
				UpdateTypes: []string{"minor", "patch"},
			},
		}
		expectedConfig := Config{
			Version: 2,
			Updates: []Dependency{
				{"github-actions", "/", schedule, cooldown, "increase", groups},
				{"docker", "/", schedule, cooldown, "increase", groups},
				{"docker", "/a", schedule, cooldown, "increase", groups},
				{"npm", "/a/b/c/d", schedule, cooldown, "increase", groups},
			},
		}
		config := Render(MatchedPaths{
			"Dockerfile":        "docker",
			"a/Dockerfile":      "docker",
			"a/b/c/d/yarn.lock": "npm",
			"/":                 "github-actions",
		})
		assert.Equal(t, expectedConfig, config)
	})
}
