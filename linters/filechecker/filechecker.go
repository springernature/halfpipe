package filechecker

import (
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"path"
)

func CheckFile(fs afero.Afero, path string, mustBeExecutable bool) error {
	if exists, err := fs.Exists(path); !exists || err != nil {
		return linterrors.NewFileError(path, "does not exist")
	}

	info, err := fs.Stat(path)
	if err != nil {
		return linterrors.NewFileError(path, "cannot be read")
	}

	if !info.Mode().IsRegular() {
		return linterrors.NewFileError(path, "is not a file")
	}

	if info.Size() == 0 {
		return linterrors.NewFileError(path, "is empty")
	}

	if mustBeExecutable && info.Mode()&0111 == 0 {
		return linterrors.NewFileError(path, "is not executable")
	}

	return nil
}

func ReadFile(fs afero.Afero, path string) (content string, err error) {
	if err = CheckFile(fs, path, false); err != nil {
		return content, err
	}

	bytez, err := fs.ReadFile(path)
	if err != nil {
		return content, err
	}

	content = string(bytez)
	return content, err
}

func GetHalfpipeFileName(fs afero.Afero, workingDir string, halfpipeFilenameOptions []string) (halfpipeFileName string, err error) {
	var foundPaths []string

	for _, p := range halfpipeFilenameOptions {
		joinedPath := path.Join(workingDir, p)
		if exists, fileNotExistErr := fs.Exists(joinedPath); exists && fileNotExistErr == nil {
			foundPaths = append(foundPaths, p)
		}
	}

	if len(foundPaths) > 1 {
		err = errors2.New(fmt.Sprintf("found %s files. Please use only 1 of those", foundPaths))
		return halfpipeFileName, err
	}

	if len(foundPaths) == 0 {
		err = linterrors.NewMissingHalfpipeFileError(halfpipeFilenameOptions)
		return halfpipeFileName, err
	}

	return foundPaths[0], nil
}
