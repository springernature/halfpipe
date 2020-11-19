package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCFDeployDefaults(t *testing.T) {
	t.Run("old apis", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}

		expected := manifest.DeployCF{
			Org:        man.Team,
			Username:   Concourse.CF.OnPrem.Username,
			Password:   Concourse.CF.OnPrem.Password,
			Manifest:   Concourse.CF.ManifestPath,
			CliVersion: Concourse.CF.Version,
		}

		assert.Equal(t, expected, deployCfDefaulter(manifest.DeployCF{}, Concourse, man))
	})

	t.Run("new apis", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}

		expected := manifest.DeployCF{
			Org:        Concourse.CF.SnPaaS.Org,
			API:        Concourse.CF.SnPaaS.API,
			Username:   Concourse.CF.SnPaaS.Username,
			Password:   Concourse.CF.SnPaaS.Password,
			TestDomain: "springernature.app",
			Manifest:   Concourse.CF.ManifestPath,
			CliVersion: Concourse.CF.Version,
		}

		assert.Equal(t, expected, deployCfDefaulter(manifest.DeployCF{API: Concourse.CF.SnPaaS.API}, Concourse, man))
	})

	t.Run("cli version", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "cf6", deployCfDefaulter(manifest.DeployCF{}, Concourse, man).CliVersion)
	})
}

func TestDoesntOverride(t *testing.T) {
	input := manifest.DeployCF{
		API:        "a",
		Org:        "b",
		Username:   "c",
		Password:   "d",
		Manifest:   "e",
		TestDomain: "f",
		CliVersion: "g",
	}

	updated := deployCfDefaulter(input, Concourse, manifest.Manifest{})

	assert.Equal(t, input, updated)
}
