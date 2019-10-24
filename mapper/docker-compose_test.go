package mapper

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingWhenFeatureToggleNotSet(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	original := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.DockerCompose{},
		},
	}

	assert.Equal(t, original, NewDockerComposeMapper(fs).Apply(original))
}

func TestConvertsDockerComposeTaskToRunTask(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	var dockerComposeContents = `
        version: 3
        services:
          some-service:
            image: appropriate/curl
            command: foo-bar`

	fs.WriteFile("docker-compose-foo.yml", []byte(dockerComposeContents), 0777)

	original := manifest.Manifest{
		FeatureToggles: []string{manifest.FeatureFlattenDockerCompose},
		Tasks: manifest.TaskList{
			manifest.DockerCompose{
				Name:          "task name",
				ManualTrigger: false,
				Vars: manifest.Vars{
					"VAR1": "VALUE1",
					"VAR2": "VALUE2",
				},
				Service:                "some-service",
				ComposeFile:            "docker-compose-foo.yml",
				SaveArtifacts:          []string{"one", "two"},
				RestoreArtifacts:       false,
				SaveArtifactsOnFailure: []string{"three", "four"},
				Retries:                0,
				NotifyOnSuccess:        false,
				Notifications: manifest.Notifications{
					OnSuccess:        []string{"#five", "#six"},
					OnSuccessMessage: "on success message",
					OnFailure:        []string{"#seven", "#eight"},
					OnFailureMessage: "on failure message",
				},
				Timeout: "1m",
			},
		},
	}

	expected := manifest.Run{
		Name:          "task name",
		ManualTrigger: false,
		Vars: manifest.Vars{
			"VAR1": "VALUE1",
			"VAR2": "VALUE2",
		},
		Docker: manifest.Docker{
			Image: "appropriate/curl",
		},
		Script:                 "foo-bar",
		SaveArtifacts:          []string{"one", "two"},
		RestoreArtifacts:       false,
		SaveArtifactsOnFailure: []string{"three", "four"},
		Retries:                0,
		NotifyOnSuccess:        false,
		Notifications: manifest.Notifications{
			OnSuccess:        []string{"#five", "#six"},
			OnSuccessMessage: "on success message",
			OnFailure:        []string{"#seven", "#eight"},
			OnFailureMessage: "on failure message",
		},
		Timeout: "1m",
	}

	assert.Equal(t, expected, NewDockerComposeMapper(fs).Apply(original).Tasks[0])
}
