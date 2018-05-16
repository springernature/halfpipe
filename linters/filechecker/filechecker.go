package filechecker

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
)

func CheckFile(fs afero.Afero, path string, mustBeExecutable bool) error {
	if exists, err := fs.Exists(path); !exists || err != nil {
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

func ReadHalfpipeFiles(fs afero.Afero, paths []string) (content string, err error) {
	var foundPaths []string

	for _, path := range paths {
		if exists, fileNotExistErr := fs.Exists(path); exists && fileNotExistErr == nil {
			foundPaths = append(foundPaths, path)
		}
	}

	if len(foundPaths) > 1 {
		err = errors.NewHalfpipeFileError(fmt.Sprintf("found %s files. Please use only 1 of those", foundPaths))
		return
	}

	if len(foundPaths) == 0 {
		err = errors.NewHalfpipeFileError(fmt.Sprintf("couldn't find any of the allowed %s files", paths))
		return
	}

	return ReadFile(fs, foundPaths[0])
}
