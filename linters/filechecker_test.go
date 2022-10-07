package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func testFs() afero.Afero {
	return afero.Afero{Fs: afero.NewMemMapFs()}
}

func TestFile_NotExists(t *testing.T) {
	fs := testFs()
	err := CheckFile(fs, "missing.file", false)

	assert.Equal(t, ErrFileNotFound.WithFile("missing.file"), err)
}

func TestFile_Empty(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0777)

	err := CheckFile(fs, ".halfpipe.io", false)
	assert.Equal(t, ErrFileEmpty.WithFile(".halfpipe.io"), err)
}

func TestFile_IsDirectory(t *testing.T) {
	fs := testFs()
	fs.Mkdir("build", 0777)

	err := CheckFile(fs, "build", false)
	assert.Equal(t, ErrFileNotAFile.WithFile("build"), err)
}

func TestFile_NotExecutable(t *testing.T) {
	fs := testFs()
	fs.WriteFile("build.sh", []byte("go test"), 0400)

	err := CheckFile(fs, "build.sh", true)
	assert.Equal(t, ErrFileNotExecutable.WithFile("build.sh"), err)

	err = CheckFile(fs, "build.sh", false)
	assert.NoError(t, err)
}

func TestFile_Happy(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)
	err := CheckFile(fs, ".halfpipe.io", true)

	assert.NoError(t, err)
}

func TestRead(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)

	content, err := ReadFile(fs, ".halfpipe.io")

	assert.NoError(t, err)
	assert.Equal(t, "foo", content)
}

func TestReadDoesCheck(t *testing.T) {
	fs := testFs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0700)
	_, err := ReadFile(fs, ".halfpipe.io")

	assert.Equal(t, ErrFileEmpty.WithFile(".halfpipe.io"), err)
}
