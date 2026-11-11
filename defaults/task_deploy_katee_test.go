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
			VelaManifest:  "vela.yaml",
			Namespace:     "katee-" + man.Team,
			CheckInterval: Actions.Katee.CheckInterval,
			MaxChecks:     Actions.Katee.MaxChecks,
			TaskBase:      manifest.TaskBase{Retries: 1},
		}

		assert.Equal(t, expected, deployKateeDefaulter(manifest.DeployKatee{}, Actions, man))
	})

	t.Run("Does not default katee namespace when set", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "Tjoho", deployKateeDefaulter(manifest.DeployKatee{Namespace: "Tjoho"}, Actions, man).Namespace)
	})

	t.Run("KATEE_NAMESPACE in vars wins over the default and the namespace field", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		vars := manifest.Vars{"KATEE_NAMESPACE": "katee-from-var"}
		assert.Equal(t, "katee-from-var", deployKateeDefaulter(manifest.DeployKatee{Vars: vars}, Actions, man).Namespace)
		assert.Equal(t, "katee-from-var", deployKateeDefaulter(manifest.DeployKatee{Vars: vars, Namespace: "Tjoho"}, Actions, man).Namespace)
	})

	t.Run("converts deploymentCheckTimeout to check_interval and max_checks", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		task := deployKateeDefaulter(manifest.DeployKatee{DeploymentCheckTimeout: 120}, Actions, man)
		assert.Equal(t, Actions.Katee.CheckInterval, task.CheckInterval)
		assert.Equal(t, 120, task.MaxChecks)
	})
}
