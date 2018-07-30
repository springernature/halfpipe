package dockercompose

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReaderWithValidFile(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(filePath, []byte(`
version: 3
services:
  db:
    vars:
    - one
    - two
    image: database/db:tag
  app:
    image: appropriate/curl
  no-image:
    build: .
`), 0777)

	expected := DockerCompose{
		Services: []Service{
			{Name: "app", Image: "appropriate/curl"},
			{Name: "db", Image: "database/db:tag"},
			{Name: "no-image", Image: ""},
		},
	}
	actual, err := NewReader(fs)()

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReaderWithoutServicesKey(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(filePath, []byte(`
app:
  image: appropriate/curl
db:
  vars:
  - one
  - two
  image: database/db:tag
no-image:
  build: .
`), 0777)

	expected := DockerCompose{
		Services: []Service{
			{Name: "app", Image: "appropriate/curl"},
			{Name: "db", Image: "database/db:tag"},
			{Name: "no-image", Image: ""},
		},
	}
	actual, err := NewReader(fs)()

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReaderWithInvalidFile(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(filePath, []byte(`
app:
  image:
  - a
  - list
db:
- image: foo
- var: bar
`), 0777)

	_, err := NewReader(fs)()

	assert.Error(t, err)
}
