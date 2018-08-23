package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
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
		Project: project.Data{
			GitURI:    "ssh@github.com:private/repo",
			GitBranch: "someBranch",
		},
	}

	man := manifest.Manifest{}
	man = manifestDefaults.Update(man)
	assert.Equal(t, manifestDefaults.RepoPrivateKey, man.Repo.PrivateKey)

	//doesn't replace existing value
	man.Repo.PrivateKey = "foo"

	man = manifestDefaults.Update(man)
	assert.Equal(t, "foo", man.Repo.PrivateKey)
	assert.Equal(t, manifestDefaults.Project.GitBranch, man.Repo.Branch)
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
			Image: config.DockerRegistry + "runImage",
		},
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, task2}}

	expectedTask2 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + "runImage",
			Username: manifestDefaults.DockerUsername,
			Password: manifestDefaults.DockerPassword,
		},
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, expectedTask2}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestDeployCfTaskWithPrePromote(t *testing.T) {

	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((gcr.private_key))",
	}

	task := manifest.DeployCF{
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
		PrePromote: []manifest.Task{
			manifest.Run{
				Script: "./blah",
				Docker: manifest.Docker{
					Image: config.DockerRegistry + "runImage",
				}},
			manifest.DockerPush{
				Image: config.DockerRegistry + "runImage",
			}},
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task}}
	expectedTask := manifest.DeployCF{
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
		PrePromote: []manifest.Task{manifest.Run{
			Script: "./blah",
			Docker: manifest.Docker{
				Image:    config.DockerRegistry + "runImage",
				Username: manifestDefaults.DockerUsername,
				Password: manifestDefaults.DockerPassword,
			}},
			manifest.DockerPush{
				Image:    config.DockerRegistry + "runImage",
				Username: manifestDefaults.DockerUsername,
				Password: manifestDefaults.DockerPassword,
			}},
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestDefaultsOnFailureTask(t *testing.T) {

	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((gcr.private_key))",
	}

	task := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image: config.DockerRegistry + "runImage",
		}}

	man := manifest.Manifest{Team: "ee", OnFailure: []manifest.Task{task}, Tasks: []manifest.Task{manifest.DockerCompose{}}}
	expectedTask := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + "runImage",
			Username: manifestDefaults.DockerUsername,
			Password: manifestDefaults.DockerPassword,
		}}

	expected := manifest.Manifest{Team: "ee", OnFailure: []manifest.Task{expectedTask}, Tasks: []manifest.Task{manifest.DockerCompose{}}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestDockerPushDefaultWhenImageIsInHalfpipeRegistry(t *testing.T) {
	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((gcr.private_key))",
	}

	imageInHalfpipeRegistry := config.DockerRegistry + "push-me"
	imageInAnotherRegistry := "some-other-registry/repo"

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{
		manifest.DockerPush{Image: imageInHalfpipeRegistry},
		manifest.DockerPush{Image: imageInAnotherRegistry},
	}}

	actual := manifestDefaults.Update(man)

	expectedTasks := manifest.TaskList{
		manifest.DockerPush{
			Username: manifestDefaults.DockerUsername,
			Password: manifestDefaults.DockerPassword,
			Image:    imageInHalfpipeRegistry,
		},
		manifest.DockerPush{
			Image: imageInAnotherRegistry,
		},
	}

	assert.Equal(t, expectedTasks, actual.Tasks)
}

func TestSetsProjectValues(t *testing.T) {
	pro := project.Data{BasePath: "foo", GitURI: "bar"}
	manifestDefaults := Defaults{
		Project: pro,
	}
	man := manifest.Manifest{}

	man = manifestDefaults.Update(man)

	assert.Equal(t, "bar", man.Repo.URI)
	assert.Equal(t, "foo", man.Repo.BasePath)
}

func TestDoesNotSetProjectValuesWhenManifestRepoUriIsSet(t *testing.T) {
	pro := project.Data{BasePath: "foo", GitURI: "bar"}
	manifestDefaults := Defaults{
		Project: pro,
	}
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com/foo/bar"

	man = manifestDefaults.Update(man)

	assert.Equal(t, "git@github.com/foo/bar", man.Repo.URI)
	assert.Equal(t, "", man.Repo.BasePath)
}

func TestSetsDefaultDockerComposeService(t *testing.T) {
	composeDefaultService := "app"
	manifestDefaults := Defaults{
		DockerComposeService: composeDefaultService,
	}

	overrideService := "asdf"

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{},
			manifest.DockerCompose{
				Service: overrideService,
			},
		},
	}
	man.Repo.URI = "git@github.com/foo/bar"

	man = manifestDefaults.Update(man)

	assert.Equal(t, composeDefaultService, man.Tasks[0].(manifest.DockerCompose).Service)
	assert.Equal(t, overrideService, man.Tasks[1].(manifest.DockerCompose).Service)
}

func TestSetsDefaultTestDomainForDeployTask(t *testing.T) {
	api := "https://api.cf"
	testDomain := "some.domain.io"
	customTestDomain := "some.other.domain.io"

	manifestDefaults := Defaults{
		CfTestDomains: map[string]string{
			api: testDomain,
		},
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{ // Well known
				API: api,
			},
			manifest.DeployCF{ // Well known but with defined testDomain
				API:        api,
				TestDomain: customTestDomain,
			},
			manifest.DeployCF{ // Unknown api
				API: "https://some.random.domain.io",
			},
		},
	}
	man = manifestDefaults.Update(man)

	assert.Equal(t, testDomain, man.Tasks[0].(manifest.DeployCF).TestDomain)
	assert.Equal(t, customTestDomain, man.Tasks[1].(manifest.DeployCF).TestDomain)
	assert.Equal(t, "", man.Tasks[2].(manifest.DeployCF).TestDomain)
}
