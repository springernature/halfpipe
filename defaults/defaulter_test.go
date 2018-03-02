package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/parser"
	"github.com/stretchr/testify/assert"
)

func TestRepoDefaultsForPublicRepo(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((github.private_key))"}

	man := parser.Manifest{Repo: parser.Repo{Uri: "https://github.com/public/repo"}}
	man = manifestDefaults.Update(man)
	assert.Empty(t, man.Repo.PrivateKey)
}

func TestRepoDefaultsForPrivateRepo(t *testing.T) {
	manifestDefaults := Defaults{RepoPrivateKey: "((github.private_key))"}

	man := parser.Manifest{Repo: parser.Repo{Uri: "ssh@github.com:private/repo"}}
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

	task1 := parser.DeployCF{}
	task2 := parser.DeployCF{
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
	}

	manifest := parser.Manifest{Team: "ee", Tasks: []parser.Task{task1, task2}}

	expectedTask1 := parser.DeployCF{
		Org:      "ee",
		Username: manifestDefaults.CfUsername,
		Password: manifestDefaults.CfPassword,
		Manifest: manifestDefaults.CfManifest,
	}

	expected := parser.Manifest{Team: "ee", Tasks: []parser.Task{expectedTask1, task2}}

	actual := manifestDefaults.Update(manifest)

	assert.Equal(t, expected, actual)
}

func TestRunTaskDefault(t *testing.T) {

	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((gcr.private_key))",
	}

	task1 := parser.Run{
		Script: "./blah",
		Docker: parser.Docker{
			Image: "Blah",
		},
	}
	task2 := parser.Run{
		Script: "./blah",
		Docker: parser.Docker{
			Image: "eu.gcr.io/halfpipe-io/runImage",
		},
	}

	manifest := parser.Manifest{Team: "ee", Tasks: []parser.Task{task1, task2}}

	expectedTask2 := parser.Run{
		Script: "./blah",
		Docker: parser.Docker{
			Image:    "eu.gcr.io/halfpipe-io/runImage",
			Username: manifestDefaults.DockerUsername,
			Password: manifestDefaults.DockerPassword,
		},
	}

	expected := parser.Manifest{Team: "ee", Tasks: []parser.Task{task1, expectedTask2}}

	actual := manifestDefaults.Update(manifest)

	assert.Equal(t, expected, actual)
}
