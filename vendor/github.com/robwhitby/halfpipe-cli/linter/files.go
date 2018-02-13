package linter

import (
	. "github.com/robwhitby/halfpipe-cli/model"
	"github.com/spf13/afero"
)

type RequiredFile struct {
	Path       string
	Executable bool
}

func requiredFiles(man Manifest) (files []RequiredFile) {
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case Run:
			files = append(files, RequiredFile{
				Path:       task.Script,
				Executable: true,
			})
		}
	}
	return
}

func LintFiles(man Manifest, fs afero.Afero) (errs []error) {
	for _, file := range requiredFiles(man) {
		if err := CheckFile(file, fs); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func CheckFile(file RequiredFile, fs afero.Afero) error {
	if exists, _ := fs.Exists(file.Path); !exists {
		return NewFileError(file.Path, "does not exist")
	}

	info, err := fs.Stat(file.Path)
	if err != nil {
		return NewFileError(file.Path, "cannot be read")
	}

	if !info.Mode().IsRegular() {
		return NewFileError(file.Path, "is not a regular file")
	}

	if info.Size() == 0 {
		return NewFileError(file.Path, "is empty")
	}

	if file.Executable && info.Mode()&0111 == 0 {
		return NewFileError(file.Path, "is not executable")
	}

	return nil
}
