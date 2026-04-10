package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	schedule := Schedule{"weekly"}
	cooldown := Cooldown{DefaultDays: 5}
	semverGroups := Groups{
		"minor-and-patch": Group{
			UpdateTypes: []string{"minor", "patch"},
		},
	}

	t.Run("Emits one entry per ecosystem regardless of how many paths matched", func(t *testing.T) {
		config := Render(MatchedPaths{
			"Dockerfile":        "docker",
			"a/Dockerfile":      "docker",
			"a/b/c/d/yarn.lock": "npm",
			"/":                 "github-actions",
		})
		assert.Equal(t, Config{
			Version: 2,
			Updates: []Dependency{
				{PackageEcosystem: "docker", Directories: []string{"/**"}, Schedule: schedule, Cooldown: cooldown},
				{PackageEcosystem: "github-actions", Directories: []string{"/"}, Schedule: schedule, Cooldown: cooldown},
				{PackageEcosystem: "npm", Directories: []string{"/**"}, Schedule: schedule, Cooldown: cooldown, VersioningStrategy: "increase", Groups: semverGroups},
			},
		}, config)
	})

	t.Run("Ecosystems are sorted alphabetically", func(t *testing.T) {
		config := Render(MatchedPaths{
			"go.mod":            "gomod",
			"package-lock.json": "npm",
			"Dockerfile":        "docker",
		})
		assert.Equal(t, []string{"docker", "gomod", "npm"}, ecosystemNames(config))
	})

	t.Run("github-actions gets directory / not /**", func(t *testing.T) {
		config := Render(MatchedPaths{"/": "github-actions"})
		assert.Equal(t, []string{"/"}, config.Updates[0].Directories)
	})
}

// ecosystemNames is a test helper that extracts the package ecosystem names from a Config.
func ecosystemNames(c Config) []string {
	names := make([]string, len(c.Updates))
	for i, dep := range c.Updates {
		names[i] = dep.PackageEcosystem
	}
	return names
}
