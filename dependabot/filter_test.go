package dependabot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("Finds Dockerfile", func(t *testing.T) {
		found := NewFilter().Filter([]string{"a", "a/b", "Dockerfile", "a/Dockerfile", "a/b/c/Dockerfile"}, []string{})
		assert.Equal(t, []string{"Dockerfile", "a/Dockerfile", "a/b/c/Dockerfile"}, found)
	})

	t.Run("Finds package-lock.json", func(t *testing.T) {
		found := NewFilter().Filter([]string{"a", "a/b", "package-lock.json", "a/package-lock.json", "a/b/c/package-lock.json"}, []string{})
		assert.Equal(t, []string{"package-lock.json", "a/package-lock.json", "a/b/c/package-lock.json"}, found)
	})

	t.Run("Finds yarn.lock", func(t *testing.T) {
		found := NewFilter().Filter([]string{"a", "a/b", "yarn.lock", "a/yarn.lock", "a/b/c/yarn.lock"}, []string{})
		assert.Equal(t, []string{"yarn.lock", "a/yarn.lock", "a/b/c/yarn.lock"}, found)
	})

	t.Run("Finds Gemfile.lock", func(t *testing.T) {
		found := NewFilter().Filter([]string{"a", "a/b", "Gemfile.lock", "a/Gemfile.lock", "a/b/c/Gemfile.lock"}, []string{})
		assert.Equal(t, []string{"Gemfile.lock", "a/Gemfile.lock", "a/b/c/Gemfile.lock"}, found)
	})

	t.Run("Skips docker and npm", func(t *testing.T) {
		found := NewFilter().Filter([]string{
			"a", "a/b", "Dockerfile", "a/Dockerfile", "a/b/c/Dockerfile",
			"a", "a/b", "package-lock.json", "a/package-lock.json", "a/b/c/package-lock.json",
			"a", "a/b", "yarn.lock", "a/yarn.lock", "a/b/c/yarn.lock",
			"a", "a/b", "Gemfile.lock", "a/Gemfile.lock", "a/b/c/Gemfile.lock",
		}, []string{"npm", "docker"})

		assert.Equal(t, []string{"Gemfile.lock", "a/Gemfile.lock", "a/b/c/Gemfile.lock"}, found)
	})
}
