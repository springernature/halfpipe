package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintUploadSLOsTask(task manifest.UploadSLOs, fs afero.Afero) (errs []error) {
	exists, err := fs.DirExists(task.Folder)
	if err != nil {
		return []error{err}
	}

	if !exists {
		return []error{ErrSLOsDirectoryNotFound(task.Folder)}
	}

	return nil
}
