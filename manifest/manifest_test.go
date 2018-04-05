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
