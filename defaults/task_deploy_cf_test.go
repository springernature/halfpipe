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
			Username:   DefaultValues.CF.OnPrem.Username,
			Password:   DefaultValues.CF.OnPrem.Password,
			Manifest:   DefaultValues.CF.ManifestPath,
			CliVersion: DefaultValues.CF.Version,
		}

		assert.Equal(t, expected, deployCfDefaulter(manifest.DeployCF{}, DefaultValues, man))
	})

	t.Run("new apis", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}

		expected := manifest.DeployCF{
			Org:        DefaultValues.CF.SnPaaS.Org,
			API:        DefaultValues.CF.SnPaaS.Api,
			Username:   DefaultValues.CF.SnPaaS.Username,
			Password:   DefaultValues.CF.SnPaaS.Password,
			TestDomain: "springernature.app",
			Manifest:   DefaultValues.CF.ManifestPath,
			CliVersion: DefaultValues.CF.Version,
		}

		assert.Equal(t, expected, deployCfDefaulter(manifest.DeployCF{API: DefaultValues.CF.SnPaaS.Api}, DefaultValues, man))
	})

	t.Run("cli version", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "cf6", deployCfDefaulter(manifest.DeployCF{}, DefaultValues, man).CliVersion)
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

	updated := deployCfDefaulter(input, DefaultValues, manifest.Manifest{})

	assert.Equal(t, input, updated)
}
