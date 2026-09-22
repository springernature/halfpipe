package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestUploadSLOsTaskWhenSpecifiedFolderDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintUploadSLOsTask(manifest.UploadSLOs{Folder: "slos"}, fs)

	assertContainsError(t, errors, ErrSLOsDirectoryNotFound("slos"))
}

func TestUploadSLOsTaskWhenSpecifiedFolderExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintUploadSLOsTask(manifest.UploadSLOs{}, fs)

	assert.Empty(t, errors, "there should be no errors")
}
