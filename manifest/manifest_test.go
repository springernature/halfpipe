package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo_UriFormats(t *testing.T) {

	type data struct {
		URI    string
		Name   string
		Public bool
	}

	var testData = []data{
		{URI: "git@github.com:private/repo", Name: "source"},
		{URI: "git@github.com:private/repo-name.git", Name: "source"},
		{URI: "git@github.com:pri-v-ate/repo.git/", Name: "source"},

		{URI: "http://github.com/private/repo-name.git/", Name: "source", Public: true},
		{URI: "https://github.com/private/repo-name", Name: "source", Public: true},
		{URI: "http://github.com/pri-v-ate/repo-name.git/", Name: "source", Public: true},
	}

	for i, test := range testData {
		repo := Repo{URI: test.URI}
		assert.Equal(t, test.Name, "source", test.Name, i)
		assert.Equal(t, test.Public, repo.IsPublic(), test.Name, i)
	}
}

func TestPipelineNameOnMaster(t *testing.T) {
	assert.Equal(t, "some-pipeline-name", actualName("some-pipeline-name", ""))
	assert.Equal(t, "some-pipeline-name", actualName("some-pipeline-name", "master"))
}

func TestPipelineNameOnBranch(t *testing.T) {
	assert.Equal(t, "some-pipeline-name-some-branch", actualName("some-pipeline-name", "some-branch"))
}

func TestPipelineNameShouldSanitizeDodgyCharactersInRepoAndBranchName(t *testing.T) {
	assert.Equal(t, "soME_pipeline-name99-some_branch", actualName(" soME$pipeline-name99 ", " some/branch "))
}

func actualName(repoName, branchName string) string {
	return Manifest{
		Pipeline: repoName,
		Triggers: TriggerList{
			GitTrigger{
				Branch: branchName,
			},
		},
	}.PipelineName()
}

func TestFlatten(t *testing.T) {
	t.Run("When its already flat", func(t *testing.T) {
		taskList := TaskList{
			DeployCF{},
			DockerPush{},
			DockerCompose{},
		}

		assert.Equal(t, taskList, taskList.Flatten())
	})

	t.Run("When its not flat", func(t *testing.T) {
		taskList := TaskList{
			DockerPush{Name: "Task 1"},
			DeployCF{
				Name: "Task 2",
				PrePromote: TaskList{
					DeployMLZip{Name: "Task 3"},
				}},
			DockerCompose{Name: "Task 4"},
			Sequence{
				Tasks: TaskList{
					Run{Name: "Task 5"},
				},
			},
			Parallel{
				Tasks: TaskList{
					Sequence{
						Tasks: TaskList{
							DeployCF{
								Name: "Task 6",
								PrePromote: TaskList{
									Run{Name: "Task 7"},
								},
							},
						},
					},
				},
			},
		}
		expected := TaskList{
			DockerPush{Name: "Task 1"},
			DeployCF{Name: "Task 2"},
			DeployMLZip{Name: "Task 3"},
			DockerCompose{Name: "Task 4"},
			Run{Name: "Task 5"},
			DeployCF{Name: "Task 6"},
			Run{Name: "Task 7"},
		}

		assert.Equal(t, expected, taskList.Flatten())
	})

}
