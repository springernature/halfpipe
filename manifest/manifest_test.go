package manifest

import (
	"testing"

	"path"

	"github.com/spf13/afero"
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

func TestReturnsErrorWhenHalfpipeDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	currentDir := "/blurg"

	_, err := ReadManifest(currentDir, fs)

	assert.Error(t, err)
}

func TestReturnsEmptyManifestWhenHalfpipeFileIsEmpty(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	currentDir := "/blurg"

	fs.WriteFile(path.Join(currentDir, ".halfpipe.io"), []byte(""), 0777)

	_, err := ReadManifest(currentDir, fs)

	assert.Error(t, err)
}

func TestReturnsErrorWhenManifestIsBroken(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	currentDir := "/blurg"

	fs.WriteFile(path.Join(currentDir, ".halfpipe.io"), []byte("WrYyYyYy"), 0777)

	_, err := ReadManifest(currentDir, fs)

	assert.Error(t, err)
}

func TestReturnsManifest(t *testing.T) {
	validHalfpipeYaml := `
team: asd
pipeline: my-pipeline
tasks:
- type: docker-compose
`
	expectedManifest := Manifest{
		Team:     "asd",
		Pipeline: "my-pipeline",
		Tasks: []Task{
			DockerCompose{},
		},
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	currentDir := "/blurg"

	fs.WriteFile(path.Join(currentDir, ".halfpipe.io"), []byte(validHalfpipeYaml), 0777)

	manifest, err := ReadManifest(currentDir, fs)

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, manifest)
}
