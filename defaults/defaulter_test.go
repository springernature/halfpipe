package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func TestRepoDefaults(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((deploy_key))"}

	man := model.Manifest{Repo: model.Repo{}}
	man = manifestDefaults.Update(man)
	assert.Equal(t, manifestDefaults.RepoPrivateKey, man.Repo.PrivateKey)

	man.Repo.PrivateKey = "foo"
	man = manifestDefaults.Update(man)
	assert.Equal(t, "foo", man.Repo.PrivateKey)
}

func TestCFDeployDefaults(t *testing.T) {

	manifestDefaults := Defaults{
		CfUsername: "((cf-credentials.username))",
		CfPassword: "((cf-credentials.password))",
		CfManifest: "manifest.yml",
		CfApiAliases: map[string]string{
			"dev":  "https://dev....com",
			"live": "https://live...com",
		},
	}

	task1 := model.DeployCF{Api: "live"}
	task2 := model.DeployCF{
		Api:      "https://doo",
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
	}

	manifest := model.Manifest{Team: "ee", Tasks: []model.Task{task1, task2}}

	expectedTask1 := model.DeployCF{
		Api:      manifestDefaults.CfApiAliases["live"],
		ApiAlias: "live",
		Org:      "ee",
		Username: manifestDefaults.CfUsername,
		Password: manifestDefaults.CfPassword,
		Manifest: manifestDefaults.CfManifest,
	}

	expected := model.Manifest{Team: "ee", Tasks: []model.Task{expectedTask1, task2}}

	actual := manifestDefaults.Update(manifest)

	assert.Equal(t, expected, actual)
}
