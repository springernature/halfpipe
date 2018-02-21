package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func TestRepoDefaultsForPublicRepo(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((github.deploy_key))"}

	man := model.Manifest{Repo: model.Repo{Uri: "https://github.com/public/repo"}}
	man = manifestDefaults.Update(man)
	assert.Empty(t, man.Repo.PrivateKey)
}

func TestRepoDefaultsForPrivateRepo(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((github.deploy_key))"}

	man := model.Manifest{Repo: model.Repo{Uri: "ssh@github.com:private/repo"}}
	man = manifestDefaults.Update(man)
	assert.Equal(t, manifestDefaults.RepoPrivateKey, man.Repo.PrivateKey)

	//doesn't replace existing value
	man.Repo.PrivateKey = "foo"
	man = manifestDefaults.Update(man)
	assert.Equal(t, "foo", man.Repo.PrivateKey)
}

func TestCFDeployDefaults(t *testing.T) {

	manifestDefaults := Defaults{
		CfUsername: "((cloudfoundry.username))",
		CfPassword: "((cloudfoundry.password))",
		CfManifest: "manifest.yml",
	}

	task1 := model.DeployCF{}
	task2 := model.DeployCF{
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
	}

	manifest := model.Manifest{Team: "ee", Tasks: []model.Task{task1, task2}}

	expectedTask1 := model.DeployCF{
		Org:      "ee",
		Username: manifestDefaults.CfUsername,
		Password: manifestDefaults.CfPassword,
		Manifest: manifestDefaults.CfManifest,
	}

	expected := model.Manifest{Team: "ee", Tasks: []model.Task{expectedTask1, task2}}

	actual := manifestDefaults.Update(manifest)

	assert.Equal(t, expected, actual)
}
