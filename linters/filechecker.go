package linters

import (
	"github.com/spf13/afero"
)

func CheckFile(fs afero.Afero, path string, mustBeExecutable bool) error {
	if exists, err := fs.Exists(path); !exists || err != nil {
		return ErrFileNotFound.WithFile(path)
	}

	info, err := fs.Stat(path)
	if err != nil {
		return ErrFileCannotRead.WithFile(path)
	}

	if !info.Mode().IsRegular() {
		return ErrFileNotAFile.WithFile(path)
	}

	if info.Size() == 0 {
		return ErrFileEmpty.WithFile(path)
	}

	if mustBeExecutable && info.Mode()&0111 == 0 {
		return ErrFileNotExecutable.WithFile(path)
	}

	return nil
}

func ReadFile(fs afero.Afero, path string) (content string, err error) {
	if err = CheckFile(fs, path, false); err != nil {
		return content, err
	}

	bytez, err := fs.ReadFile(path)
	if err != nil {
		return content, ErrFileInvalid.WithFile(path)
	}

	content = string(bytez)
	return content, nil
}
