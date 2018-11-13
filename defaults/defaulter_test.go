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
			GitURI: "ssh@github.com:private/repo",
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
		CfUsername:  "((cloudfoundry.username))",
		CfPassword:  "((cloudfoundry.password))",
		CfManifest:  "manifest.yml",
		CfApiSnPaas: "((snpaas-api))",
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

func TestCFDeployDefaultsForSNPaaS(t *testing.T) {

	manifestDefaults := Defaults{
		CfUsernameSnPaas: "u",
		CfPasswordSnPaas: "p",
		CfOrgSnPaas:      "o",
		CfApiSnPaas:      "a",
	}

	task := manifest.DeployCF{
		API: "a",
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task}}

	expectedTask := manifest.DeployCF{
		API:      "a",
		Org:      manifestDefaults.CfOrgSnPaas,
		Username: manifestDefaults.CfUsernameSnPaas,
		Password: manifestDefaults.CfPasswordSnPaas,
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected, actual)
}

func TestRunTaskDockerDefault(t *testing.T) {

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

	expectedTask2Docker := manifest.Docker{
		Image:    config.DockerRegistry + "runImage",
		Username: manifestDefaults.DockerUsername,
		Password: manifestDefaults.DockerPassword,
	}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, task1.Docker, actual.Tasks[0].(manifest.Run).Docker)
	assert.Equal(t, expectedTask2Docker, actual.Tasks[1].(manifest.Run).Docker)
}

func TestDockerComposeTaskSaveArtifactsOnFailureDefault(t *testing.T) {
	composeDefaultService := "app"

	manifestDefaults := Defaults{
		Project: project.Data{
			HalfpipeFilePath: ".halfpipe.io.yaml",
		},
		DockerComposeService: composeDefaultService,
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{},
		},
	}

	expectedDCTask := manifest.DockerCompose{
		SaveArtifactsOnFailure: []string{manifestDefaults.Project.HalfpipeFilePath},
		Parallel:               false,
		Retries:                0,
	}

	man = manifestDefaults.Update(man)


	actual := manifestDefaults.Update(man)

	assert.Equal(t, expectedDCTask.SaveArtifactsOnFailure, actual.Tasks[0].(manifest.DockerCompose).SaveArtifactsOnFailure)
}

func TestRunTaskSaveArtifactsOnFailureDefault(t *testing.T) {

	manifestDefaults := Defaults{
		Project: project.Data{
			HalfpipeFilePath: ".halfpipe.io.yaml",
		},
	}

	task1 := manifest.DockerCompose{
		Name:                   "",
		Command:                "",
		ManualTrigger:          false,
		Vars:                   nil,
		Service:                "",
		SaveArtifacts:          nil,
		RestoreArtifacts:       false,
		SaveArtifactsOnFailure: nil,
		Parallel:               false,
		Retries:                0,
	}

	task2 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image: "Blah",
		},
		SaveArtifactsOnFailure: []string{".halfpipe.io.yml"},
	}

	task3 := manifest.Run{
		Script: "./blah",
		Docker: manifest.Docker{
			Image: "Blah",
		},
		SaveArtifactsOnFailure: []string{"build.sh", "gradle.file"},
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, task2, task3}}

	expectedRunTask := manifest.Run{
		Script:                 "./blah",
		Docker:                 manifest.Docker{Image: "Blah"},
		SaveArtifactsOnFailure: []string{manifestDefaults.Project.HalfpipeFilePath},
	}

	expectedRunTask2 := manifest.Run{
		Script:                 "./blah",
		Docker:                 manifest.Docker{Image: "Blah"},
		SaveArtifactsOnFailure: []string{"build.sh", "gradle.file", manifestDefaults.Project.HalfpipeFilePath},
	}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expectedRunTask.SaveArtifactsOnFailure, actual.Tasks[0].(manifest.Run).SaveArtifactsOnFailure)
	assert.Equal(t, task2.SaveArtifactsOnFailure, actual.Tasks[1].(manifest.Run).SaveArtifactsOnFailure)
	assert.Equal(t, expectedRunTask2.SaveArtifactsOnFailure, actual.Tasks[2].(manifest.Run).SaveArtifactsOnFailure)
}

func TestDeployCfTaskWithPrePromote(t *testing.T) {
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
				},
				SaveArtifactsOnFailure: []string{".halfpipe.io"}},

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
		PrePromote: []manifest.Task{
			manifest.Run{
				Script: "./blah",
				Docker: manifest.Docker{
					Image:    config.DockerRegistry + "runImage",
					Username: DefaultValues.DockerUsername,
					Password: DefaultValues.DockerPassword,
				},
				Vars: map[string]string{
					"ARTIFACTORY_USERNAME": "((artifactory.username))",
					"ARTIFACTORY_PASSWORD": "((artifactory.password))",
					"ARTIFACTORY_URL":      "((artifactory.url))",
				},
				SaveArtifactsOnFailure: []string{".halfpipe.io"},
			},
			manifest.DockerPush{
				Image:    config.DockerRegistry + "runImage",
				Username: DefaultValues.DockerUsername,
				Password: DefaultValues.DockerPassword,
				Vars: map[string]string{
					"ARTIFACTORY_USERNAME": "((artifactory.username))",
					"ARTIFACTORY_PASSWORD": "((artifactory.password))",
					"ARTIFACTORY_URL":      "((artifactory.url))",
				},
			},
		},
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask}}

	actual := DefaultValues.Update(man)

	assert.Equal(t, expected, actual)
}

func TestDockerPushDefaultWhenImageIsInHalfpipeRegistry(t *testing.T) {
	imageInHalfpipeRegistry := config.DockerRegistry + "push-me"
	imageInAnotherRegistry := "some-other-registry/repo"

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{
		manifest.DockerPush{Image: imageInHalfpipeRegistry},
		manifest.DockerPush{Image: imageInAnotherRegistry},
	}}

	actual := DefaultValues.Update(man)

	expectedTasks := manifest.TaskList{
		manifest.DockerPush{
			Username: DefaultValues.DockerUsername,
			Password: DefaultValues.DockerPassword,
			Image:    imageInHalfpipeRegistry,
			Vars: map[string]string{
				"ARTIFACTORY_USERNAME": "((artifactory.username))",
				"ARTIFACTORY_PASSWORD": "((artifactory.password))",
				"ARTIFACTORY_URL":      "((artifactory.url))",
			},
		},
		manifest.DockerPush{
			Image: imageInAnotherRegistry,
			Vars: map[string]string{
				"ARTIFACTORY_USERNAME": "((artifactory.username))",
				"ARTIFACTORY_PASSWORD": "((artifactory.password))",
				"ARTIFACTORY_URL":      "((artifactory.url))",
			},
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
	assert.Equal(t, pro.BasePath, man.Repo.BasePath)
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

func TestSetsArtifactoryUsernameAndPassword(t *testing.T) {
	otherUsername := "someOtherUsername"
	otherPassword := "someOtherPassword"

	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerCompose{},
			manifest.DockerPush{},
			manifest.DeployCF{
				PrePromote: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{},
					manifest.DockerCompose{},
				},
			},
			manifest.Run{
				Vars: map[string]string{
					"ARTIFACTORY_USERNAME": otherUsername,
					"ARTIFACTORY_PASSWORD": otherPassword,
				},
			},
			manifest.ConsumerIntegrationTest{},
		},
	}

	updated := DefaultValues.Update(man)

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[0].(manifest.Run).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[0].(manifest.Run).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[0].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, otherUsername, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, otherPassword, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryUrl, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_URL"])
}
