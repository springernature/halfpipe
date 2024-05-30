package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestKateeDeployDefaults(t *testing.T) {
	t.Run("katee - actions", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf", Platform: "actions"}

		expected := manifest.DeployKatee{
			VelaManifest:    "vela.yaml",
			Tag:             "version",
			Namespace:       "katee-" + man.Team,
			Environment:     "asdf",
			PlatformVersion: "v1",
		}

		assert.Equal(t, expected, deployKateeDefaulter(manifest.DeployKatee{}, Actions, man))
	})

	t.Run("katee - update-pipeline-enabled - concourse", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf", FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}
		assert.Equal(t, "version", deployKateeDefaulter(manifest.DeployKatee{}, Actions, man).Tag)
	})

	t.Run("katee - update-pipeline-disabled - concourse", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "gitref", deployKateeDefaulter(manifest.DeployKatee{}, Actions, man).Tag)
	})

	t.Run("Does not default katee namespace when set", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "Tjoho", deployKateeDefaulter(manifest.DeployKatee{Namespace: "Tjoho"}, Actions, man).Namespace)
	})

	t.Run("Does not default katee env when set", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "blurgh", deployKateeDefaulter(manifest.DeployKatee{Environment: "blurgh"}, Actions, man).Environment)
	})

	t.Run("Does not override platform_version", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf", Platform: "actions"}

		input := manifest.DeployKatee{PlatformVersion: "v1337"}

		assert.Equal(t, "v1337", deployKateeDefaulter(input, Actions, man).PlatformVersion)

	})
}
