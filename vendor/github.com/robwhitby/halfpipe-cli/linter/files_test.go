package linter

import (
	"testing"

	. "github.com/robwhitby/halfpipe-cli/model"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func fs() afero.Afero {
	return afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/some/root")}
}

func TestFile_NotExists(t *testing.T) {
	fs := fs()
	err := CheckFile(RequiredFile{Path: "missing.file"}, fs)

	assert.Equal(t, NewFileError("missing.file", "does not exist"), err)
}

func TestFile_Empty(t *testing.T) {
	fs := fs()
	fs.WriteFile(".halfpipe.io", []byte{}, 0777)

	err := CheckFile(RequiredFile{Path: ".halfpipe.io"}, fs)
	assert.Equal(t, NewFileError(".halfpipe.io", "is empty"), err)
}

func TestFile_IsDirectory(t *testing.T) {
	fs := fs()
	fs.Mkdir("build", 0777)

	err := CheckFile(RequiredFile{Path: "build"}, fs)
	assert.Equal(t, NewFileError("build", "is not a regular file"), err)
}

func TestFile_NotExecutable(t *testing.T) {
	fs := fs()
	fs.WriteFile("build.sh", []byte("go test"), 0400)

	err := CheckFile(RequiredFile{Path: "build.sh", Executable: true}, fs)
	assert.Equal(t, NewFileError("build.sh", "is not executable"), err)

	err = CheckFile(RequiredFile{Path: "build.sh", Executable: false}, fs)
	assert.Nil(t, err)
}

func TestFile_Happy(t *testing.T) {
	fs := fs()
	fs.WriteFile(".halfpipe.io", []byte("foo"), 0700)
	err := CheckFile(RequiredFile{Path: ".halfpipe.io", Executable: true}, fs)

	assert.Nil(t, err)
}

func TestRequiredFiles_FindAllFiles(t *testing.T) {
	manifest := Manifest{
		Team: "ee",
		Repo: Repo{Uri: "http://github.com/foo/bar.git"},
		Tasks: []Task{
			Run{
				Script: "./build1.sh",
				Image:  "alpine",
			},
			Run{
				Script: "./build2.sh",
				Image:  "alpine",
			}},
	}

	files := requiredFiles(manifest)

	expected := []RequiredFile{{
		Path:       "./build1.sh",
		Executable: true,
	}, {
		Path:       "./build2.sh",
		Executable: true,
	}}

	assert.Equal(t, expected, files)
}

func TestLintFiles(t *testing.T) {
	fs := fs()
	manifest := Manifest{
		Team: "ee",
		Repo: Repo{Uri: "http://github.com/foo/bar.git"},
		Tasks: []Task{Run{
			Script: "./build1.sh",
			Image:  "alpine",
		}},
	}

	errors := LintFiles(manifest, fs)
	assert.Equal(t, []error{NewFileError("./build1.sh", "does not exist")}, errors)
}
