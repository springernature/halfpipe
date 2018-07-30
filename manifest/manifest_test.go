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

func TestManifest_HasDockerComposeTask(t *testing.T) {
	tests := []struct {
		manifest Manifest
		result   bool
	}{
		{
			manifest: Manifest{Tasks: TaskList{Run{}}},
			result:   false,
		},
		{
			manifest: Manifest{Tasks: TaskList{Run{}, DockerCompose{}, DeployCF{}}},
			result:   true,
		},
		{
			manifest: Manifest{OnFailure: TaskList{Run{}, DockerCompose{}, DeployCF{}}},
			result:   true,
		},
		{
			manifest: Manifest{Tasks: TaskList{DeployCF{}}},
			result:   false,
		},
		{
			manifest: Manifest{Tasks: TaskList{DeployCF{PrePromote: TaskList{Run{}, DockerCompose{}}}}},
			result:   true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.result, test.manifest.HasDockerComposeTask())
	}
}
