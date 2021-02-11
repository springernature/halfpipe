package defaults

import (
	"errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGit(t *testing.T) {
	dummyGitBranchResolver := func() (string, error) {
		return "", nil
	}

	t.Run("does nothing when URI is not set", func(t *testing.T) {
		trigger := manifest.GitTrigger{}
		assert.Equal(t, trigger, defaultGitTrigger(trigger, Concourse, dummyGitBranchResolver, manifest.Platform("concourse")))
	})

	t.Run("private repos", func(t *testing.T) {
		defaults := Concourse
		defaults.Project = project.Data{
			GitURI: "ssh@github.com:private/repo",
		}

		t.Run("no private key set", func(t *testing.T) {
			trigger := manifest.GitTrigger{}
			assert.Equal(t, defaults.RepoPrivateKey, defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("concourse")).PrivateKey)
		})

		t.Run("private key set", func(t *testing.T) {
			privateKey := "sup"
			trigger := manifest.GitTrigger{PrivateKey: privateKey}
			assert.Equal(t, privateKey, defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("concourse")).PrivateKey)
		})
	})

	t.Run("public repos", func(t *testing.T) {

		t.Run("http", func(t *testing.T) {
			defaults := Concourse
			defaults.Project = project.Data{
				GitURI: "http://github.com/springernature/halfpipe.git",
			}

			trigger := manifest.GitTrigger{}
			updatedTrigger := defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("concourse"))

			assert.Equal(t, "git@github.com:springernature/halfpipe.git", updatedTrigger.URI)
			assert.Equal(t, defaults.RepoPrivateKey, updatedTrigger.PrivateKey)
		})

		t.Run("https", func(t *testing.T) {
			defaults := Concourse
			defaults.Project = project.Data{
				GitURI: "https://github.com/springernature/halfpipe.git",
			}

			trigger := manifest.GitTrigger{}
			updatedTrigger := defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("concourse"))

			assert.Equal(t, "git@github.com:springernature/halfpipe.git", updatedTrigger.URI)
			assert.Equal(t, defaults.RepoPrivateKey, updatedTrigger.PrivateKey)
		})

	})

	t.Run("project values", func(t *testing.T) {
		defaults := Concourse
		defaults.Project = project.Data{BasePath: "foo", GitURI: "bar"}

		expectedTrigger := manifest.GitTrigger{
			PrivateKey: defaults.RepoPrivateKey,
			URI:        "bar",
			BasePath:   "foo",
		}

		assert.Equal(t, expectedTrigger, defaultGitTrigger(manifest.GitTrigger{}, defaults, dummyGitBranchResolver, manifest.Platform("concourse")))
	})

	t.Run("does not overwrite URI when set", func(t *testing.T) {
		defaults := Concourse
		defaults.Project = project.Data{BasePath: "foo", GitURI: "bar"}

		trigger := manifest.GitTrigger{
			URI: "git@github.com/foo/bar",
		}

		updated := defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("concourse"))

		assert.Equal(t, "git@github.com/foo/bar", updated.URI)
		assert.Equal(t, "foo", updated.BasePath)
	})

	t.Run("branch", func(t *testing.T) {
		t.Run("Does nothing when its set", func(t *testing.T) {
			trigger := manifest.GitTrigger{
				Branch: "kehe",
			}
			assert.Equal(t, "kehe", defaultGitTrigger(trigger, Concourse, dummyGitBranchResolver, manifest.Platform("concourse")).Branch)
		})

		t.Run("Defaults to master when on the master branch", func(t *testing.T) {
			gitBranchResolver := func() (string, error) {
				return "master", nil
			}

			trigger := manifest.GitTrigger{
				Branch: "",
			}
			assert.Equal(t, "master", defaultGitTrigger(trigger, Concourse, gitBranchResolver, manifest.Platform("concourse")).Branch)
		})

		t.Run("Defaults to main when on the main branch", func(t *testing.T) {
			gitBranchResolver := func() (string, error) {
				return "main", nil
			}

			trigger := manifest.GitTrigger{}
			assert.Equal(t, "main", defaultGitTrigger(trigger, Concourse, gitBranchResolver, manifest.Platform("concourse")).Branch)
		})

		t.Run("Does nothing when on a random branch and branch is not set", func(t *testing.T) {
			gitBranchResolver := func() (string, error) {
				return "RaNdOm BrAnCh", nil
			}

			trigger := manifest.GitTrigger{}

			assert.Equal(t, "", defaultGitTrigger(trigger, Concourse, gitBranchResolver, manifest.Platform("concourse")).Branch)
		})

		t.Run("Does nothing if the branch resolver fails", func(t *testing.T) {
			gitBranchResolver := func() (string, error) {
				return "", errors.New("random failure. this is not an issue since we catch these errors down the line in the git trigger linter")
			}

			trigger := manifest.GitTrigger{}
			assert.Equal(t, "", defaultGitTrigger(trigger, Concourse, gitBranchResolver, manifest.Platform("concourse")).Branch)
		})

	})

	t.Run("actions", func(t *testing.T) {
		defaults := Actions
		defaults.Project = project.Data{
			GitURI: "ssh@github.com:private/repo",
		}
		trigger := manifest.GitTrigger{}
		assert.Equal(t, trigger, defaultGitTrigger(trigger, defaults, dummyGitBranchResolver, manifest.Platform("actions")))
	})
}
