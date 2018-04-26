package manifest

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfHalfpipeFileAlreadyExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte(""), 777)

	sampleGenerator := NewSampleGenerator(fs)

	err := sampleGenerator.Generate()

	assert.Equal(t, err, ErrHalfpipeAlreadyExists)
}

func TestWritesSample(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	sampleGenerator := NewSampleGenerator(fs)

	err := sampleGenerator.Generate()

	assert.Nil(t, err)

	bytes, err := fs.ReadFile(".halfpipe.io")
	assert.Nil(t, err)

	expected := `team: CHANGE-ME
pipeline: CHANGE-ME
tasks:
- type: run
  name: CHANGE-ME OPTIONAL NAME IN CONCOURSE UI
  manual_trigger: false
  script: ./gradlew CHANGE-ME
  docker:
    image: CHANGE-ME:tag
`
	assert.Equal(t, string(bytes), expected)

}
