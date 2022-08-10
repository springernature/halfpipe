package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKateeDeployDefaults(t *testing.T) {
	t.Run("katee - actions", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf", Platform: "actions"}

		expected := manifest.DeployKatee{
			VelaManifest: "vela.yaml",
			Tag:          "version",
		}

		assert.Equal(t, expected, deployKateeDefaulter(manifest.DeployKatee{}, Actions, man))
	})

	t.Run("katee - update-pipeline-enabled - concourse", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf", FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}

		expected := manifest.DeployKatee{
			VelaManifest: "vela.yaml",
			Tag:          "version",
		}

		assert.Equal(t, expected, deployKateeDefaulter(manifest.DeployKatee{}, Actions, man))
	})

	t.Run("katee - update-pipeline-disabled - concourse", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}

		expected := manifest.DeployKatee{
			VelaManifest: "vela.yaml",
			Tag:          "gitref",
		}

		assert.Equal(t, expected, deployKateeDefaulter(manifest.DeployKatee{}, Actions, man))
	})
}
