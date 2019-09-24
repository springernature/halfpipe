package pipeline

import (
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersHttpGitResource(t *testing.T) {
	gitURI := "git@github.com:springernature/repo.git"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: gitURI,
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
					"branch": "master",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersSshGitResource(t *testing.T) {
	gitURI := "git@github.com:springernature/repo.git/"
	privateKey := "blurgh"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:        gitURI,
				PrivateKey: privateKey,
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
					"branch":      "master",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithWatchesAndIgnores(t *testing.T) {
	gitURI := "git@github.com:springernature/repo.git/"
	privateKey := "blurgh"
	watches := []string{"watch1", "watch2"}
	ignores := []string{"ignore1", "ignore2"}

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:          gitURI,
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
					"branch":       "master",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersHttpGitResourceWithGitCrypt(t *testing.T) {
	gitURI := "git@github.com:springernature/repo.git"
	gitCrypt := "AABBFF66"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:         gitURI,
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
					"branch":        "master",
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(man))
}

func TestRendersGitResourceWithBranchIfSet(t *testing.T) {
	gitURI := "git@github.com:springernature/repo.git"
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
