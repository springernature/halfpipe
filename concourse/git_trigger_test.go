package concourse

import (
	"github.com/springernature/halfpipe/config"
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersHttpGitResource(t *testing.T) {
	gitURI := "git@blah.com:springernature/repo.git"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:    gitURI,
				Branch: "main",
			},
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: gitName,
				Type: "git",
				Source: atc.Source{
					"uri":    gitURI,
					"branch": "main",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersSshGitResource(t *testing.T) {
	gitURI := "git@blah.com:springernature/repo.git/"
	privateKey := "blurgh"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:        gitURI,
				PrivateKey: privateKey,
				Branch:     "main",
			},
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: gitName,
				Type: "git",
				Source: atc.Source{
					"uri":         gitURI,
					"private_key": privateKey,
					"branch":      "main",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithWatchesAndIgnores(t *testing.T) {
	gitURI := "git@blah.com:springernature/repo.git/"
	privateKey := "blurgh"
	watches := []string{"watch1", "watch2"}
	ignores := []string{"ignore1", "ignore2"}

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:          gitURI,
				Branch:       "main",
				PrivateKey:   privateKey,
				WatchedPaths: watches,
				IgnoredPaths: ignores,
			},
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: gitName,
				Type: "git",
				Source: atc.Source{
					"uri":          gitURI,
					"private_key":  privateKey,
					"paths":        watches,
					"ignore_paths": ignores,
					"branch":       "main",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersHttpGitResourceWithGitCrypt(t *testing.T) {
	gitURI := "git@blah.com:springernature/repo.git"
	gitCrypt := "AABBFF66"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:         gitURI,
				Branch:      "main",
				GitCryptKey: gitCrypt,
			},
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: gitName,
				Type: "git",
				Source: atc.Source{
					"uri":           gitURI,
					"git_crypt_key": gitCrypt,
					"branch":        "main",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithBranchIfSet(t *testing.T) {
	gitURI := "git@blah.com:springernature/repo.git"
	branch := "master"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:    gitURI,
				Branch: branch,
			},
		},
	}

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: gitName,
				Type: "git",
				Source: atc.Source{
					"uri":    gitURI,
					"branch": branch,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersTasksWithDepth1IfShallowIsSet(t *testing.T) {
	taskName := "runTask"
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				Shallow: true,
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Name: taskName,
			},
		},
	}

	rendered := testPipeline().Render(man)

	task, _ := rendered.Jobs.Lookup(taskName)
	assert.Equal(t, "git", (task.Plan[0].InParallel.Steps)[0].Get)
	assert.Equal(t, 1, (task.Plan[0].InParallel.Steps)[0].Params["depth"])
}

func TestRenderWithGitTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "t1", Script: "asd.sh"},
			manifest.DeployCF{Name: "t2", ManualTrigger: true},
			manifest.DockerPush{Name: "t3"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	getGitStep := (config.Jobs[0].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, gitName, getGitStep.Name())
	assert.True(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[1].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[0].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[2].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[1].Name, getGitStep.Passed[0])
	assert.True(t, getGitStep.Trigger)
}

func TestRenderWithGitManualTrigger(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				ManualTrigger: true,
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "t1", Script: "asd.sh"},
			manifest.DeployCF{Name: "t2", ManualTrigger: true},
			manifest.DockerPush{Name: "t3"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	getGitStep := (config.Jobs[0].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, gitName, getGitStep.Name())
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[1].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[0].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[2].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[1].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)
}

func TestRendersWebHookAssistedGitResources(t *testing.T) {
	gitURIs := map[string]string{
		"git@github.com/foo/repo.git":                "",
		config.WebHookAssistedGitPrefix + "repo.git": webHookAssistedResourceCheckInterval,
	}
	for uri, expectedInterval := range gitURIs {
		t.Run(uri, func(t *testing.T) {
			man := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.GitTrigger{
						URI: uri,
					},
				},
			}
			resource, _ := testPipeline().Render(man).Resources.Lookup(gitName)
			assert.Equal(t, expectedInterval, resource.CheckEvery)
		})
	}
}
