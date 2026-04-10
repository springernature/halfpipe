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

	t.Run("Ecosystems with registries emit registry definitions and references", func(t *testing.T) {
		config := Render(MatchedPaths{
			"pom.xml": "maven",
			"go.mod":  "gomod",
		})
		// gomod has no registry; maven references ee-artifactory
		gomod := config.Updates[0]
		assert.Equal(t, "gomod", gomod.PackageEcosystem)
		assert.Nil(t, gomod.Registries)

		maven := config.Updates[1]
		assert.Equal(t, "maven", maven.PackageEcosystem)
		assert.Equal(t, []string{"sn-artifactory"}, maven.Registries)

		// top-level registries block is populated
		assert.Equal(t, map[string]Registry{
			"sn-artifactory": registryDefinitions["sn-artifactory"],
		}, config.Registries)
	})

	t.Run("No registries block when no ecosystem needs one", func(t *testing.T) {
		config := Render(MatchedPaths{"go.mod": "gomod"})
		assert.Nil(t, config.Registries)
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
