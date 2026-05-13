package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	schedule := Schedule{"weekly"}
	cooldown := Cooldown{DefaultDays: 5}

	commitMessage := CommitMessage{Prefix: "chore", Include: "scope"}

	t.Run("Emits one entry per ecosystem with discovered directories", func(t *testing.T) {
		config := Render(MatchedPaths{
			"Dockerfile":        "docker",
			"a/Dockerfile":      "docker",
			"a/b/c/d/yarn.lock": "npm",
			"/":                 "github-actions",
		})
		assert.Equal(t, Config{
			Version: 2,
			Updates: []Dependency{
				{PackageEcosystem: "docker", Directories: []string{"/", "/a"}, Schedule: schedule, Cooldown: cooldown, OpenPullRequestsLimit: 10, Labels: []string{"dependencies", "docker"}, CommitMessage: commitMessage, Groups: allGroup},
				{PackageEcosystem: "github-actions", Directories: []string{"/"}, Schedule: schedule, Cooldown: cooldown, OpenPullRequestsLimit: 10, Labels: []string{"dependencies", "github-actions"}, CommitMessage: commitMessage, Groups: allGroup, Ignore: []Ignore{{DependencyName: "springernature/*", UpdateTypes: []string{"version-update:semver-minor", "version-update:semver-patch"}}}},
				{PackageEcosystem: "npm", Directories: []string{"/a/b/c/d"}, Schedule: schedule, Cooldown: cooldown, OpenPullRequestsLimit: 10, Labels: []string{"dependencies", "npm"}, CommitMessage: commitMessage, VersioningStrategy: "increase", Groups: semverGroups},
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

	t.Run("Maven has no registries", func(t *testing.T) {
		config := Render(MatchedPaths{
			"pom.xml": "maven",
		})
		maven := config.Updates[0]
		assert.Equal(t, "maven", maven.PackageEcosystem)
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
