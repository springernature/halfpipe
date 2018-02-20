package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo_UriFormats(t *testing.T) {

	type data struct {
		Uri         string
		Name        string
		Public      bool
		GitCryptKey string
	}

	var testData = []data{
		{Uri: "git@github.com:private/repo", Name: "repo"},
		{Uri: "git@github.com:private/repo-name.git", Name: "repo-name"},
		{Uri: "git@github.com:pri-v-ate/repo.git/", Name: "repo"},

		{Uri: "http://github.com/private/repo-name.git/", Name: "repo-name", Public: true},
		{Uri: "https://github.com/private/repo-name", Name: "repo-name", Public: true},
		{Uri: "http://github.com/pri-v-ate/repo-name.git/", Name: "repo-name", Public: true},
	}

	for i, test := range testData {
		repo := Repo{Uri: test.Uri}
		assert.Equal(t, test.Name, repo.GetName(), test.Name, i)
		assert.Equal(t, test.Public, repo.IsPublic(), test.Name, i)
	}
}
