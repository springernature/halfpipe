package samplegenerator

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/project"
	"github.com/stretchr/testify/assert"
)

type FakeProjectResolver struct {
	p   project.Data
	err error
}

func (pr FakeProjectResolver) Parse(workingDir string, ignoreMissingHalfpipeFile bool, halfpipeFilenameOptions []string) (p project.Data, err error) {
	return pr.p, pr.err
}

func TestFailsIfHalfpipeFileAlreadyExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io.yml", []byte(""), 0777)

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

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Data{RootName: "myApp"}}, "/home/user/src/myApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io.yml")
	assert.Nil(t, err)

	expected := `# yaml-language-server: $schema=https://github.com/springernature/halfpipe/releases/latest/download/schema.json
team: <team name>
pipeline: myApp
platform: concourse
tasks:
- type: run
  name: <task name>
  script: <script>
  docker:
    image: <image:tag>
`
	assert.Equal(t, expected, string(bytes))
}

func TestWritesSampleWhenExecutedInASubDirectory(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Data{
		BasePath: "subApp",
		RootName: "myApp",
		GitURI:   "",
	}}, "/home/user/src/myApp/subApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io.yml")
	assert.Nil(t, err)

	expected := `# yaml-language-server: $schema=https://github.com/springernature/halfpipe/releases/latest/download/schema.json
team: <team name>
pipeline: myApp-subApp
platform: concourse
triggers:
- type: git
  watched_paths:
  - subApp
tasks:
- type: run
  name: <task name>
  script: <script>
  docker:
    image: <image:tag>
`
	assert.Equal(t, expected, string(bytes))
}

func TestWritesSampleWhenExecutedInASubSubDirectory(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs, FakeProjectResolver{p: project.Data{
		BasePath: "subFolder/subApp",
		RootName: "myApp",
		GitURI:   "",
	}}, "/home/user/src/myApp/subFolder/subApp")

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io.yml")
	assert.Nil(t, err)

	expected := `# yaml-language-server: $schema=https://github.com/springernature/halfpipe/releases/latest/download/schema.json
team: <team name>
pipeline: myApp-subFolder-subApp
platform: concourse
triggers:
- type: git
  watched_paths:
  - subFolder/subApp
tasks:
- type: run
  name: <task name>
  script: <script>
  docker:
    image: <image:tag>
`
	assert.Equal(t, expected, string(bytes))
}
