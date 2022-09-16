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
				{"docker", "/", schedule},
				{"docker", "/a", schedule},
				{"npm", "/a/b/c/d", schedule},
				{"github-actions", "/", schedule},
			},
		}
		config := NewRender().Render([]string{"Dockerfile", "a/Dockerfile", "a/b/c/d/yarn.lock", "github-actions"})
		assert.Equal(t, expectedConfig, config)
	})
}
