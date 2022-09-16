package dependabot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	filterWithNoSkips := NewFilter([]string{})

	t.Run("matches paths to ecosystems", func(t *testing.T) {
		found := filterWithNoSkips.Filter([]string{"a", "a/b", "Dockerfile", "a/Dockerfile", "package-lock.json", "a/b/c/Gemfile.lock"})
		assert.Equal(t, MatchedPaths{
			"Dockerfile":         "docker",
			"a/Dockerfile":       "docker",
			"package-lock.json":  "npm",
			"a/b/c/Gemfile.lock": "bundler",
		}, found)
	})

	t.Run("skips docker and npm", func(t *testing.T) {
		found := NewFilter([]string{"npm", "docker"}).Filter([]string{
			"a", "a/b", "Dockerfile", "a/Dockerfile", "a/b/c/Dockerfile",
			"a", "a/b", "package-lock.json", "a/package-lock.json", "a/b/c/package-lock.json",
			"a", "a/b", "yarn.lock", "a/yarn.lock", "a/b/c/yarn.lock",
			"a", "a/b", "Gemfile.lock", "a/Gemfile.lock", "a/b/c/Gemfile.lock",
		})

		assert.Equal(t, MatchedPaths{"Gemfile.lock": "bundler", "a/Gemfile.lock": "bundler", "a/b/c/Gemfile.lock": "bundler"}, found)
	})

	t.Run("matches once for all github workflows", func(t *testing.T) {
		found := filterWithNoSkips.Filter([]string{
			"a", "a/b", ".hidden/asd", ".github/workflows/codeql.yml", ".github/workflows/someOtherWorklow.yml",
		})

		assert.Equal(t, MatchedPaths{"/": "github-actions"}, found)

	})
}
