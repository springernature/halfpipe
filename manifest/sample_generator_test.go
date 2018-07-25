package manifest

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe/project"
	"github.com/pkg/errors"
)

type FakeProjectResolver struct {
	p   project.Project
	err error
}

func (pr FakeProjectResolver) Parse(workingDir string) (p project.Project, err error) {
	return pr.p, pr.err
}

func TestFailsIfHalfpipeFileAlreadyExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(""), 777)

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{}, "/some/path")

	err := sampleGenerator.Generate()

	assert.Equal(t, err, ErrHalfpipeAlreadyExists)
}

func TestFailsIfProjectResolverErrorsOut(t *testing.T) {
	expectedError := errors.New("Oeh noes")

	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{err: expectedError}, "/some/path")

	err := sampleGenerator.Generate()

	assert.Equal(t, err, expectedError)
}

func TestWritesSample(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{}, "/some/path")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: CHANGE-ME
tasks:
- type: run
  name: CHANGE-ME OPTIONAL NAME IN CONCOURSE UI
  script: ./gradlew CHANGE-ME
  docker:
    image: CHANGE-ME:tag
`
	assert.Equal(t, string(bytes), expected)
}

func TestWritesSampleWhenExecutedInASubDirectory(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Project{
		BasePath: "subApp",
		GitURI:   "",
	}}, "/path/to/repo/subApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: CHANGE-ME
repo:
  watched_paths:
  - subApp
tasks:
- type: run
  name: CHANGE-ME OPTIONAL NAME IN CONCOURSE UI
  script: ./gradlew CHANGE-ME
  docker:
    image: CHANGE-ME:tag
`
	assert.Equal(t, string(bytes), expected)
}
