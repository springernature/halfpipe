package shell

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShell_Render_SadPath(t *testing.T) {

	t.Run("task doesn't exist", func(t *testing.T) {
		renderer := New("task name that doesn't exist")
		actual, err := renderer.Render(manifest.Manifest{Tasks: manifest.TaskList{manifest.Run{Name: "task name"}}})
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("task exists but type not supported", func(t *testing.T) {
		renderer := New("task name")
		actual, err := renderer.Render(manifest.Manifest{Tasks: manifest.TaskList{manifest.DockerPush{Name: "task name"}}})
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

}
