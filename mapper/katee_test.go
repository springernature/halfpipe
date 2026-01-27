package mapper

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestKateeMapper_ReturnsErrorWhenManifestCannotBeOpened(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	mapper := NewKateeMapper(fs)

	_, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		manifest.DeployKatee{
			VelaManifest: "nonexistent.yaml",
		},
	}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.yaml")
}

func TestKateeMapper_ParsesManifestAndPopulatesKateeManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	velaContent := `kind: Application
metadata:
  name: "hello"
  namespace: "default"
spec:
  components:
    - name: my-app
      type: snstateless
      properties:
        image: nginx:latest
`
	fs.WriteFile("vela.yaml", []byte(velaContent), 0644)

	mapper := NewKateeMapper(fs)
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		manifest.DeployKatee{
			VelaManifest: "vela.yaml",
		},
	}})

	assert.NoError(t, err)
	deployKatee := updated.Tasks[0].(manifest.DeployKatee)
	assert.Equal(t, "Application", deployKatee.KateeManifest.Kind)
	assert.Equal(t, "hello", deployKatee.KateeManifest.Metadata.Name)
	assert.Equal(t, "default", deployKatee.KateeManifest.Metadata.Namespace)

}

func TestKateeMapper_HandlesParallelAndSequenceTasks(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	velaContent := `kind: Application
spec:
  components:
    - name: nested-app
      type: worker
`
	fs.WriteFile("vela.yaml", []byte(velaContent), 0644)

	mapper := NewKateeMapper(fs)
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.DeployKatee{
					VelaManifest: "vela.yaml",
				},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.DeployKatee{
							VelaManifest: "vela.yaml",
						},
					},
				},
			},
		},
	}})

	assert.NoError(t, err)

	parallel := updated.Tasks[0].(manifest.Parallel)
	deployKatee1 := parallel.Tasks[0].(manifest.DeployKatee)
	assert.Equal(t, "Application", deployKatee1.KateeManifest.Kind)

	sequence := parallel.Tasks[1].(manifest.Sequence)
	deployKatee2 := sequence.Tasks[0].(manifest.DeployKatee)
	assert.Equal(t, "Application", deployKatee2.KateeManifest.Kind)
}

func TestKateeMapper_ReturnsErrorForInvalidYAML(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml", []byte("invalid: yaml: content:"), 0644)

	mapper := NewKateeMapper(fs)
	_, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		manifest.DeployKatee{
			VelaManifest: "vela.yaml",
		},
	}})

	assert.Error(t, err)
}
