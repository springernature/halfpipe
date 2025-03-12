package mapper

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"path"
	"strconv"
	"testing"
)

func TestReturnsErrorWhenTheManifestCannotBeOpened(t *testing.T) {
	mapper := NewCfMapper()
	path := "somePath.txt"
	_, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		manifest.DeployCF{
			Manifest: path,
		},
	}})

	assert.Equal(t, fmt.Sprintf("open %s: no such file or directory", path), err.Error())
}

func TestMapsTasks(t *testing.T) {
	normalDeployPath := path.Join("/tmp", strconv.Itoa(rand.Int()))
	normalDeploy := `---
applications:
- name: asd
`

	dockerManifestPath := path.Join("/tmp", strconv.Itoa(rand.Int()))
	dockerManifest := `---
applications:
- name: wryy
  docker:
    image: nginx
`
	cfWithNormalPush := manifest.DeployCF{
		Manifest: normalDeployPath,
	}
	cfWithDockerPush := manifest.DeployCF{
		Manifest: dockerManifestPath,
	}

	fs := afero.Afero{Fs: afero.NewOsFs()}

	fs.WriteFile(normalDeployPath, []byte(normalDeploy), 0777)
	fs.WriteFile(dockerManifestPath, []byte(dockerManifest), 0777)
	defer fs.Remove(normalDeployPath)
	defer fs.Remove(dockerManifestPath)

	mapper := NewCfMapper()
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		cfWithNormalPush,
		cfWithDockerPush,
	}})

	assert.NoError(t, err)
	assert.False(t, updated.Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.Equal(t, updated.Tasks[0].(manifest.DeployCF).CfApplication.Name, "asd")

	assert.True(t, updated.Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.Equal(t, updated.Tasks[1].(manifest.DeployCF).CfApplication.Name, "wryy")
}

func TestMapsTasksForSeqAndParallel(t *testing.T) {
	normalDeployPath := path.Join("/tmp", strconv.Itoa(rand.Int()))
	normalDeploy := `---
applications:
- name: asd
`

	dockerManifestPath := path.Join("/tmp", strconv.Itoa(rand.Int()))
	dockerManifest := `---
applications:
- name: wryy
  docker:
    image: nginx
`
	cfWithNormalPush := manifest.DeployCF{
		Manifest: normalDeployPath,
	}
	cfWithDockerPush := manifest.DeployCF{
		Manifest: dockerManifestPath,
	}

	fs := afero.Afero{Fs: afero.NewOsFs()}

	fs.WriteFile(normalDeployPath, []byte(normalDeploy), 0777)
	fs.WriteFile(dockerManifestPath, []byte(dockerManifest), 0777)
	defer fs.Remove(normalDeployPath)
	defer fs.Remove(dockerManifestPath)

	mapper := NewCfMapper()
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		cfWithNormalPush,
		cfWithDockerPush,
		manifest.Parallel{
			Tasks: manifest.TaskList{
				cfWithNormalPush,
				cfWithDockerPush,
				manifest.Sequence{
					Tasks: manifest.TaskList{
						cfWithNormalPush,
						cfWithDockerPush,
					},
				},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						cfWithNormalPush,
						cfWithDockerPush,
					},
				},
			},
		},
	}})

	assert.NoError(t, err)

	assert.False(t, updated.Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[1].(manifest.DeployCF).IsDockerPush)

	assert.False(t, updated.Tasks[2].(manifest.Parallel).Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[2].(manifest.Parallel).Tasks[1].(manifest.DeployCF).IsDockerPush)

	assert.False(t, updated.Tasks[2].(manifest.Parallel).Tasks[3].(manifest.Sequence).Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[2].(manifest.Parallel).Tasks[3].(manifest.Sequence).Tasks[1].(manifest.DeployCF).IsDockerPush)

}
