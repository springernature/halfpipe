package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo_UriFormats(t *testing.T) {

	type data struct {
		URI         string
		Name        string
		Public      bool
		GitCryptKey string
	}

	var testData = []data{
		{URI: "git@github.com:private/repo", Name: "repo"},
		{URI: "git@github.com:private/repo-name.git", Name: "repo-name"},
		{URI: "git@github.com:pri-v-ate/repo.git/", Name: "repo"},

		{URI: "http://github.com/private/repo-name.git/", Name: "repo-name", Public: true},
		{URI: "https://github.com/private/repo-name", Name: "repo-name", Public: true},
		{URI: "http://github.com/pri-v-ate/repo-name.git/", Name: "repo-name", Public: true},
	}

	for i, test := range testData {
		repo := Repo{URI: test.URI}
		assert.Equal(t, test.Name, repo.GetName(), test.Name, i)
		assert.Equal(t, test.Public, repo.IsPublic(), test.Name, i)
	}
}
