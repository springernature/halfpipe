package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestBuildpackDefaults(t *testing.T) {
	t.Run("builder", func(t *testing.T) {
		expected := manifest.Buildpack{
			Builder: Concourse.Buildpack.Builder,
		}
		assert.Equal(t, expected, buildpackDefaulter(manifest.Buildpack{}, Concourse))
	})
}
