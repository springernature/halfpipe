package defaults

import (
	"path"
	"testing"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
)

func TestTriggers(t *testing.T) {
	t.Run("GitTrigger", func(t *testing.T) {
		t.Run("does nothing when URI is not set", func(t *testing.T) {
			manifestDefaults := Defaults{RepoPrivateKey: "((halfpipe-github.private_key))"}

			trigger := manifest.GitTrigger{}
			man := manifest.Manifest{
				Triggers: manifest.TriggerList{
					trigger,
				},
			}
			man = manifestDefaults.Update(man)
			assert.Equal(t, trigger, man.Triggers[0])
		})

		t.Run("private repos", func(t *testing.T) {
			manifestDefaults := Defaults{
				RepoPrivateKey: "((halfpipe-github.private_key))",
				Project: project.Data{
					GitURI: "ssh@github.com:private/repo",
				},
			}

			t.Run("no private key set", func(t *testing.T) {
				man := manifest.Manifest{
					Triggers: manifest.TriggerList{
						manifest.GitTrigger{},
					},
				}
				man = manifestDefaults.Update(man)
				assert.Equal(t, manifestDefaults.RepoPrivateKey, man.Triggers[0].(manifest.GitTrigger).PrivateKey)
			})

			t.Run("private key set", func(t *testing.T) {
				//doesn't replace existing value
				man := manifest.Manifest{
					Triggers: manifest.TriggerList{
						manifest.GitTrigger{
							PrivateKey: "foo",
						},
					},
				}

				man = manifestDefaults.Update(man)
				assert.Equal(t, "foo", man.Triggers[0].(manifest.GitTrigger).PrivateKey)
			})
		})

		t.Run("project values", func(t *testing.T) {
			pro := project.Data{BasePath: "foo", GitURI: "bar"}
			manifestDefaults := Defaults{
				Project: pro,
			}

			expectedManifest := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.GitTrigger{
						URI:      "bar",
						BasePath: "foo",
					},
				},
			}

			assert.Equal(t, expectedManifest.Triggers, manifestDefaults.Update(manifest.Manifest{}).Triggers)
		})

		t.Run("does not overwrite URI when set", func(t *testing.T) {
			pro := project.Data{BasePath: "foo", GitURI: "bar"}
			manifestDefaults := Defaults{
				Project: pro,
			}
			man := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.GitTrigger{
						URI: "git@github.com/foo/bar",
					},
				},
			}

			man = manifestDefaults.Update(man)

			assert.Equal(t, "git@github.com/foo/bar", man.Triggers[0].(manifest.GitTrigger).URI)
			assert.Equal(t, "foo", man.Triggers[0].(manifest.GitTrigger).BasePath)
		})
	})

	t.Run("DockerTrigger", func(t *testing.T) {
		t.Run("does not do anything when the image is not from our registry", func(t *testing.T) {
			man := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.CronTrigger{},
					manifest.DockerTrigger{
						Image: "ubuntu",
					},
				},
			}

			manifestDefaults := Defaults{DockerUsername: "meehp0", DockerPassword: "maahp"}
			assert.Equal(t, man.Triggers[1], manifestDefaults.Update(man).Triggers[1])
		})

		t.Run("sets the username and password if not set when using private registry", func(t *testing.T) {

			manifestDefaults := Defaults{DockerUsername: "meehp0", DockerPassword: "maahp"}

			image := path.Join(config.DockerRegistry, "ubuntu")
			man := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.CronTrigger{},
					manifest.DockerTrigger{
						Image: image,
					},
				},
			}

			expectedTrigger := manifest.DockerTrigger{
				Image:    image,
				Username: manifestDefaults.DockerUsername,
				Password: manifestDefaults.DockerPassword,
			}

			assert.Equal(t, expectedTrigger, manifestDefaults.Update(man).Triggers[1])
		})
	})
}

func TestCFDeployDefaults(t *testing.T) {

	manifestDefaults := Defaults{
		CfUsername:  "((cloudfoundry.username))",
		CfPassword:  "((cloudfoundry.password))",
		CfManifest:  "manifest.yml",
		CfAPISnPaas: "((snpaas-api))",
	}

	task1 := manifest.DeployCF{}
	task2 := manifest.DeployCF{
		Name:     "deploy to org space",
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task1, task2}}

	expectedTask1 := manifest.DeployCF{
		Name:     "deploy-cf",
		Org:      "ee",
		Username: manifestDefaults.CfUsername,
		Password: manifestDefaults.CfPassword,
		Manifest: manifestDefaults.CfManifest,
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask1, task2}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected.Tasks, actual.Tasks)
}

func TestCFDeployDefaultsForSNPaaS(t *testing.T) {

	manifestDefaults := Defaults{
		CfUsernameSnPaas: "u",
		CfPasswordSnPaas: "p",
		CfOrgSnPaas:      "o",
		CfAPISnPaas:      "a",
	}

	task := manifest.DeployCF{
		API: "a",
	}

	man := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{task}}

	expectedTask := manifest.DeployCF{
		Name:     "deploy-cf",
		API:      "a",
		Org:      manifestDefaults.CfOrgSnPaas,
		Username: manifestDefaults.CfUsernameSnPaas,
		Password: manifestDefaults.CfPasswordSnPaas,
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask}}

	actual := manifestDefaults.Update(man)

	assert.Equal(t, expected.Tasks, actual.Tasks)
}

func TestRunTaskDockerDefault(t *testing.T) {

	manifestDefaults := Defaults{
		DockerUsername: "_json_key",
		DockerPassword: "((halfpipe-gcr.private_key))",
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
		Name:     "deploy-cf",
		Org:      "org",
		Space:    "space",
		Username: "user",
		Password: "pass",
		Manifest: "man.yml",
		PrePromote: []manifest.Task{
			manifest.Run{
				Name:   "run blah",
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
				Timeout:                DefaultValues.Timeout,
			},
			manifest.DockerPush{
				Name:     "docker-push",
				Image:    config.DockerRegistry + "runImage",
				Username: DefaultValues.DockerUsername,
				Password: DefaultValues.DockerPassword,
				Vars: map[string]string{
					"ARTIFACTORY_USERNAME": "((artifactory.username))",
					"ARTIFACTORY_PASSWORD": "((artifactory.password))",
					"ARTIFACTORY_URL":      "((artifactory.url))",
				},
				Timeout:        DefaultValues.Timeout,
				DockerfilePath: "Dockerfile",
			},
		},
		Timeout: DefaultValues.Timeout,
	}

	expected := manifest.Manifest{Team: "ee", Tasks: []manifest.Task{expectedTask}}

	actual := DefaultValues.Update(man)

	assert.Equal(t, expected.Tasks, actual.Tasks)
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
			Name:     "docker-push",
			Username: DefaultValues.DockerUsername,
			Password: DefaultValues.DockerPassword,
			Image:    imageInHalfpipeRegistry,
			Vars: map[string]string{
				"ARTIFACTORY_USERNAME": "((artifactory.username))",
				"ARTIFACTORY_PASSWORD": "((artifactory.password))",
				"ARTIFACTORY_URL":      "((artifactory.url))",
			},
			Timeout:        DefaultValues.Timeout,
			DockerfilePath: "Dockerfile",
		},
		manifest.DockerPush{
			Name:  "docker-push (1)",
			Image: imageInAnotherRegistry,
			Vars: map[string]string{
				"ARTIFACTORY_USERNAME": "((artifactory.username))",
				"ARTIFACTORY_PASSWORD": "((artifactory.password))",
				"ARTIFACTORY_URL":      "((artifactory.url))",
			},
			Timeout:        DefaultValues.Timeout,
			DockerfilePath: "Dockerfile",
		},
	}

	assert.Equal(t, expectedTasks, actual.Tasks)
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
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[0].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[1].(manifest.DockerCompose).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[2].(manifest.DockerPush).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[3].(manifest.DeployCF).PrePromote[0].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[3].(manifest.DeployCF).PrePromote[1].(manifest.DockerPush).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[3].(manifest.DeployCF).PrePromote[2].(manifest.DockerCompose).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, otherUsername, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, otherPassword, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[4].(manifest.Run).Vars["ARTIFACTORY_URL"])

	assert.Equal(t, DefaultValues.ArtifactoryUsername, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_USERNAME"])
	assert.Equal(t, DefaultValues.ArtifactoryPassword, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_PASSWORD"])
	assert.Equal(t, DefaultValues.ArtifactoryURL, updated.Tasks[5].(manifest.ConsumerIntegrationTest).Vars["ARTIFACTORY_URL"])
}

func TestSetsTimeout(t *testing.T) {
	timeout := "5m"

	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerCompose{Timeout: timeout},
			manifest.DockerPush{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.DockerPush{},
							manifest.DockerCompose{Timeout: timeout},
						},
					},
					manifest.Run{},
				},
			},
			manifest.ConsumerIntegrationTest{},
			manifest.DeployMLModules{},
			manifest.DeployMLZip{},
		},
	}

	updated := DefaultValues.Update(man)

	assert.Equal(t, DefaultValues.Timeout, updated.Tasks[0].GetTimeout())
	assert.Equal(t, timeout, updated.Tasks[1].GetTimeout())
	assert.Equal(t, DefaultValues.Timeout, updated.Tasks[2].GetTimeout())

	// CF with prepromote
	cfTask := (updated.Tasks[3].(manifest.Parallel).Tasks[0]).(manifest.DeployCF)
	assert.Equal(t, DefaultValues.Timeout, cfTask.GetTimeout())
	assert.Equal(t, DefaultValues.Timeout, cfTask.PrePromote[0].GetTimeout())
	assert.Equal(t, DefaultValues.Timeout, cfTask.PrePromote[1].GetTimeout())
	assert.Equal(t, timeout, cfTask.PrePromote[2].GetTimeout())

	runTask := (updated.Tasks[3].(manifest.Parallel).Tasks[1]).(manifest.Run)
	assert.Equal(t, DefaultValues.Timeout, runTask.GetTimeout())
	assert.Equal(t, DefaultValues.Timeout, updated.Tasks[5].GetTimeout())
	assert.Equal(t, DefaultValues.Timeout, updated.Tasks[6].GetTimeout())
}

func TestSetsNames(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{URI: "https://github.com/springernature/halfpipe.git"},
		Tasks: []manifest.Task{
			manifest.Run{Script: "asd.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Script: "asd.sh"},
					manifest.Run{Name: "test", Script: "asd.sh"},
					manifest.Run{Name: "test", Script: "asd.sh"},
				},
			},
			manifest.Run{Script: "asd.sh"},
			manifest.Run{Name: "test", Script: "asd.sh"},
			manifest.Run{Name: "test", Script: "fgh.sh"},
			manifest.DeployCF{
				Name: "deploy-cf",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test", Script: "asd.sh"},
					manifest.Run{Script: "asd.sh"},
					manifest.Run{Name: "test", Script: "asd.sh"},
					manifest.Run{Script: "asd.sh"},
				},
			},
			manifest.DeployCF{
				Name: "deploy-cf",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test", Script: "asd.sh"},
					manifest.Run{Script: "asd.sh"},
					manifest.Run{Name: "test", Script: "asd.sh"},
					manifest.Run{Script: "asd.sh"},
				},
			},
			manifest.DeployCF{},
			manifest.DockerPush{},
			manifest.DockerPush{},
			manifest.DockerPush{},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DockerPush{Name: "push to docker hub"},
			manifest.DockerPush{Name: "push to docker hub"},
		},
	}

	expededWithoutAllTheOtherFields := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Name: "run asd.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "run asd.sh (1)"},
					manifest.Run{Name: "test"},
					manifest.Run{Name: "test (1)"},
				},
			},
			manifest.Run{Name: "run asd.sh (2)"},
			manifest.Run{Name: "test (2)"},
			manifest.Run{Name: "test (3)"},
			manifest.DeployCF{
				Name: "deploy-cf",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test"},
					manifest.Run{Name: "run asd.sh"},
					manifest.Run{Name: "test (1)"},
					manifest.Run{Name: "run asd.sh (1)"},
				},
			},
			manifest.DeployCF{
				Name: "deploy-cf (1)",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test"},
					manifest.Run{Name: "run asd.sh"},
					manifest.Run{Name: "test (1)"},
					manifest.Run{Name: "run asd.sh (1)"},
				},
			},
			manifest.DeployCF{Name: "deploy-cf (2)"},
			manifest.DockerPush{Name: "docker-push"},
			manifest.DockerPush{Name: "docker-push (1)"},
			manifest.DockerPush{Name: "docker-push (2)"},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DeployCF{Name: "deploy to dev (1)"},
			manifest.DockerPush{Name: "push to docker hub"},
			manifest.DockerPush{Name: "push to docker hub (1)"},
		},
	}

	updated := DefaultValues.Update(man)

	assert.Len(t, expededWithoutAllTheOtherFields.Tasks, len(updated.Tasks))
	for i, updatedTask := range updated.Tasks {
		if updateParallelTask, isParallelTask := updatedTask.(manifest.Parallel); isParallelTask {
			expectedParallelTask := expededWithoutAllTheOtherFields.Tasks[i].(manifest.Parallel)
			for pi, pTask := range updateParallelTask.Tasks {
				assert.Equal(t, expectedParallelTask.Tasks[pi].GetName(), pTask.GetName())
			}
		} else {
			assert.Equal(t, expededWithoutAllTheOtherFields.Tasks[i].GetName(), updatedTask.GetName())
			if updatedDeployCf, isDeployCf := updatedTask.(manifest.DeployCF); isDeployCf {
				expectedDeployCf := expededWithoutAllTheOtherFields.Tasks[i].(manifest.DeployCF)
				for ppi, ppTask := range updatedDeployCf.PrePromote {
					assert.Equal(t, expectedDeployCf.PrePromote[ppi].GetName(), ppTask.GetName())
				}
			}
		}
	}
}

func TestAddsAUpdateTaskIfUpdateFeatureIsSet(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: []string{manifest.FeatureUpdatePipeline},
		Tasks: manifest.TaskList{
			manifest.Run{Script: "asd.sh"},
		},
	}

	expected := manifest.TaskList{
		manifest.Update{
			Timeout: "1h",
		},
		manifest.Run{
			Name:   "run asd.sh",
			Script: "asd.sh",
			Vars: map[string]string{
				"ARTIFACTORY_USERNAME": "((artifactory.username))",
				"ARTIFACTORY_PASSWORD": "((artifactory.password))",
				"ARTIFACTORY_URL":      "((artifactory.url))",
			},
			Timeout: "1h",
		},
	}

	updated := DefaultValues.Update(man)
	assert.Equal(t, expected, updated.Tasks)
}
