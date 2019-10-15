package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGit(t *testing.T) {
	t.Run("does nothing when URI is not set", func(t *testing.T) {
		trigger := manifest.GitTrigger{}
		assert.Equal(t, trigger, defaultGitTrigger(trigger, DefaultValuesNew))
	})

	t.Run("private repos", func(t *testing.T) {
		defaults := DefaultValuesNew
		defaults.Project = project.Data{
			GitURI: "ssh@github.com:private/repo",
		}

		t.Run("no private key set", func(t *testing.T) {
			trigger := manifest.GitTrigger{}
			assert.Equal(t, defaults.RepoPrivateKey, defaultGitTrigger(trigger, defaults).PrivateKey)
		})

		t.Run("private key set", func(t *testing.T) {
			privateKey := "sup"
			trigger := manifest.GitTrigger{PrivateKey: privateKey}
			assert.Equal(t, privateKey, defaultGitTrigger(trigger, defaults).PrivateKey)
		})
	})

	t.Run("public repos", func(t *testing.T) {

		t.Run("http", func(t *testing.T) {
			defaults := DefaultValuesNew
			defaults.Project = project.Data{
				GitURI: "http://github.com/springernature/halfpipe.git",
			}

			trigger := manifest.GitTrigger{}
			updatedTrigger := defaultGitTrigger(trigger, defaults)

			assert.Equal(t, "git@github.com:springernature/halfpipe.git", updatedTrigger.URI)
			assert.Equal(t, defaults.RepoPrivateKey, updatedTrigger.PrivateKey)
		})

		t.Run("https", func(t *testing.T) {
			defaults := DefaultValuesNew
			defaults.Project = project.Data{
				GitURI: "https://github.com/springernature/halfpipe.git",
			}

			trigger := manifest.GitTrigger{}
			updatedTrigger := defaultGitTrigger(trigger, defaults)

			assert.Equal(t, "git@github.com:springernature/halfpipe.git", updatedTrigger.URI)
			assert.Equal(t, defaults.RepoPrivateKey, updatedTrigger.PrivateKey)
		})

	})

	t.Run("project values", func(t *testing.T) {
		defaults := DefaultValuesNew
		defaults.Project = project.Data{BasePath: "foo", GitURI: "bar"}

		expectedTrigger := manifest.GitTrigger{
			PrivateKey: defaults.RepoPrivateKey,
			URI:        "bar",
			BasePath:   "foo",
		}

		assert.Equal(t, expectedTrigger, defaultGitTrigger(manifest.GitTrigger{}, defaults))
	})

	t.Run("does not overwrite URI when set", func(t *testing.T) {
		defaults := DefaultValuesNew
		defaults.Project = project.Data{BasePath: "foo", GitURI: "bar"}

		trigger := manifest.GitTrigger{
			URI: "git@github.com/foo/bar",
		}

		updated := defaultGitTrigger(trigger, defaults)

		assert.Equal(t, "git@github.com/foo/bar", updated.URI)
		assert.Equal(t, "foo", updated.BasePath)
	})
}
