package file_checker

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
)

func CheckFile(fs afero.Afero, path string, mustBeExecutable bool) error {
	if exists, _ := fs.Exists(path); !exists {
		return errors.NewFileError(path, "does not exist")
	}

	info, err := fs.Stat(path)
	if err != nil {
		return errors.NewFileError(path, "cannot be read")
	}

	if !info.Mode().IsRegular() {
		return errors.NewFileError(path, "is not a file")
	}

	if info.Size() == 0 {
		return errors.NewFileError(path, "is empty")
	}

	if mustBeExecutable && info.Mode()&0111 == 0 {
		return errors.NewFileError(path, "is not executable")
	}

	return nil
}

func ReadFile(fs afero.Afero, path string) (content string, err error) {
	if err = CheckFile(fs, path, false); err != nil {
		return
	}

	bytez, err := fs.ReadFile(path)
	if err != nil {
		return
	}

	content = string(bytez)
	return
}
