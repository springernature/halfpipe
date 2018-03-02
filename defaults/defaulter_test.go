package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRepoDefaultsForPublicRepo(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((github.private_key))"}

	man := manifest.Manifest{}
	man = manifestDefaults.Update(man)
	assert.Empty(t, man.Repo.PrivateKey)
}

func TestRepoDefaultsForPrivateRepo(t *testing.T) {
	manifestDefaults := Defaults{
		RepoPrivateKey: "((github.private_key))",
		Project: Project{
			GitUri: "ssh@github.com:private/repo",
		},
	}

	man := manifest.Manifest{}
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

	task1 := manifest.DeployCF{}
	task2 := manifest.DeployCF{
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, task2}}

	expectedTask1 := manifest.DeployCF{
		Org:      "ee",
		Username: manifestDefaults.CfUsername,
		Password: manifestDefaults.CfPassword,
		Manifest: manifestDefaults.CfManifest,
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask1, task2}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestRunTaskDefault(t *testing.T) {

	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((gcr.private_key))",
	}

	task1 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image: "Blah",
		},
	}
	task2 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image: "eu.gcr.io/halfpipe-io/runImage",
		},
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, task2}}

	expectedTask2 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/runImage",
			Username: manifestDefaults.DockerUsername,
			Password: manifestDefaults.DockerPassword,
		},
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, expectedTask2}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestSetsProjectValues(t *testing.T) {
	project := Project{BasePath: "foo", GitUri: "bar"}
	manifestDefaults := Defaults{
		Project: project,
	}
	man := manifest.Manifest{}

	man = manifestDefaults.Update(man)

	assert.Equal(t, "bar", man.Repo.Uri)
	assert.Equal(t, "foo", man.Repo.BasePath)
}
