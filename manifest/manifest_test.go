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
	assert.Equal(t, "some-pipeline-name", Manifest{Pipeline: "some-pipeline-name"}.PipelineName())
	assert.Equal(t, "some-pipeline-name", Manifest{Pipeline: "some-pipeline-name", Repo: Repo{Branch: "master"}}.PipelineName())
}

func TestPipelineNameOnBranch(t *testing.T) {
	assert.Equal(t, "some-pipeline-name-some-branch", Manifest{Pipeline: "some-pipeline-name", Repo: Repo{Branch: "some-branch"}}.PipelineName())
}
