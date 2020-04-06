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
	mapper := NewCFDockerPushMapper()
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
	cfWithGeneratedManifest := manifest.DeployCF{
		Manifest: "../../manifest.yml",
	}

	fs := afero.Afero{Fs: afero.NewOsFs()}

	fs.WriteFile(normalDeployPath, []byte(normalDeploy), 0777)
	fs.WriteFile(dockerManifestPath, []byte(dockerManifest), 0777)
	defer fs.Remove(normalDeployPath)
	defer fs.Remove(dockerManifestPath)

	mapper := NewCFDockerPushMapper()
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		cfWithNormalPush,
		cfWithDockerPush,
		cfWithGeneratedManifest,
	}})

	assert.NoError(t, err)
	assert.False(t, updated.Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.False(t, updated.Tasks[2].(manifest.DeployCF).IsDockerPush)
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
	cfWithGeneratedManifest := manifest.DeployCF{
		Manifest: "../../manifest.yml",
	}

	fs := afero.Afero{Fs: afero.NewOsFs()}

	fs.WriteFile(normalDeployPath, []byte(normalDeploy), 0777)
	fs.WriteFile(dockerManifestPath, []byte(dockerManifest), 0777)
	defer fs.Remove(normalDeployPath)
	defer fs.Remove(dockerManifestPath)

	mapper := NewCFDockerPushMapper()
	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{
		cfWithNormalPush,
		cfWithDockerPush,
		cfWithGeneratedManifest,
		manifest.Parallel{
			Tasks: manifest.TaskList{
				cfWithNormalPush,
				cfWithDockerPush,
				cfWithGeneratedManifest,
				manifest.Sequence{
					Tasks: manifest.TaskList{
						cfWithNormalPush,
						cfWithDockerPush,
						cfWithGeneratedManifest,
					},
				},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						cfWithNormalPush,
						cfWithDockerPush,
						cfWithGeneratedManifest,
					},
				},
			},
		},
	}})

	assert.NoError(t, err)

	assert.False(t, updated.Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.False(t, updated.Tasks[2].(manifest.DeployCF).IsDockerPush)

	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[3].(manifest.Parallel).Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[2].(manifest.DeployCF).IsDockerPush)

	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[3].(manifest.Sequence).Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[3].(manifest.Parallel).Tasks[3].(manifest.Sequence).Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[3].(manifest.Sequence).Tasks[2].(manifest.DeployCF).IsDockerPush)

	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[4].(manifest.Sequence).Tasks[0].(manifest.DeployCF).IsDockerPush)
	assert.True(t, updated.Tasks[3].(manifest.Parallel).Tasks[4].(manifest.Sequence).Tasks[1].(manifest.DeployCF).IsDockerPush)
	assert.False(t, updated.Tasks[3].(manifest.Parallel).Tasks[4].(manifest.Sequence).Tasks[2].(manifest.DeployCF).IsDockerPush)
}
