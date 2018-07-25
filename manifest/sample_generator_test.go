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

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{}, "/home/user/src/myApp")

	err := sampleGenerator.Generate()

	assert.Equal(t, err, ErrHalfpipeAlreadyExists)
}

func TestFailsIfProjectResolverErrorsOut(t *testing.T) {
	expectedError := errors.New("Oeh noes")

	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{err: expectedError}, "/home/user/src/myApp")

	err := sampleGenerator.Generate()

	assert.Equal(t, err, expectedError)
}

func TestWritesSample(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Project{RootName: "myApp"}}, "/home/user/src/myApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: myApp
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
		RootName: "myApp",
		GitURI:   "",
	}}, "/home/user/src/myApp/subApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: myApp-subApp
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

func TestWritesSampleWhenExecutedInASubSubDirectory(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Project{
		BasePath: "subFolder/subApp",
		RootName: "myApp",
		GitURI:   "",
	}}, "/home/user/src/myApp/subFolder/subApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: myApp-subFolder-subApp
repo:
  watched_paths:
  - subFolder/subApp
tasks:
- type: run
  name: CHANGE-ME OPTIONAL NAME IN CONCOURSE UI
  script: ./gradlew CHANGE-ME
  docker:
    image: CHANGE-ME:tag
`
	assert.Equal(t, string(bytes), expected)
}
