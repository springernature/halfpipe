package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCFDeployDefaults(t *testing.T) {
	t.Run("snpaas", func(t *testing.T) {
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

		assert.Equal(t, expected, deployCfDefaulter(manifest.DeployCF{}, Concourse, man))
	})

	t.Run("cli version", func(t *testing.T) {
		man := manifest.Manifest{Team: "asdf"}
		assert.Equal(t, "cf7", deployCfDefaulter(manifest.DeployCF{}, Concourse, man).CliVersion)
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
