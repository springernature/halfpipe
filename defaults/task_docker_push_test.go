package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestWhenPublicImage(t *testing.T) {
	task := manifest.DockerPush{Image: "asdf", DockerfilePath: "something", Tag: "git"}

	assert.Equal(t, task, dockerPushDefaulter(task, manifest.Manifest{}, DefaultValues))
}

func TestPrivateImage(t *testing.T) {
	task := manifest.DockerPush{Image: path.Join(config.DockerRegistry, "push-me"), DockerfilePath: "something"}

	expected := manifest.DockerPush{
		Image:          path.Join(config.DockerRegistry, "push-me"),
		DockerfilePath: "something",
		Username:       DefaultValues.Docker.Username,
		Password:       DefaultValues.Docker.Password,
		Tag:            "gitref",
	}

	assert.Equal(t, expected, dockerPushDefaulter(task, manifest.Manifest{}, DefaultValues))
}

func TestSetsTheDockerFilePath(t *testing.T) {
	assert.Equal(t, "Dockerfile", dockerPushDefaulter(manifest.DockerPush{}, manifest.Manifest{}, DefaultValues).DockerfilePath)
}

func TestTag(t *testing.T) {
	t.Run("when pipeline isn't versioned", func(t *testing.T) {
		t.Run("when tag is empty, tag defaults to git", func(t *testing.T) {
			expected := manifest.DockerPush{
				DockerfilePath: "Dockerfile",
				Tag:            "gitref",
			}

			assert.Equal(t, expected, dockerPushDefaulter(manifest.DockerPush{}, manifest.Manifest{}, DefaultValues))
		})

		t.Run("when tag is set, it does nothing", func(t *testing.T) {
			tag := "NotAThingWillBeCauthByLinter"
			expected := manifest.DockerPush{
				DockerfilePath: "Dockerfile",
				Tag:            tag,
			}

			assert.Equal(t, expected, dockerPushDefaulter(manifest.DockerPush{Tag: tag}, manifest.Manifest{}, DefaultValues))

		})
	})

	t.Run("when pipeline is versioned", func(t *testing.T) {
		man := manifest.Manifest{
			FeatureToggles: []string{
				manifest.FeatureUpdatePipeline,
			},
		}

		t.Run("when tag is empty, tag defaults to version", func(t *testing.T) {

			expected := manifest.DockerPush{
				DockerfilePath: "Dockerfile",
				Tag:            "version",
			}

			assert.Equal(t, expected, dockerPushDefaulter(manifest.DockerPush{}, man, DefaultValues))
		})

		t.Run("when tag is set, it uses it", func(t *testing.T) {
			tag := "NotAThingWillBeCauthByLinter"
			expected := manifest.DockerPush{
				DockerfilePath: "Dockerfile",
				Tag:            tag,
			}

			assert.Equal(t, expected, dockerPushDefaulter(manifest.DockerPush{Tag: tag}, man, DefaultValues))
		})
	})

}
