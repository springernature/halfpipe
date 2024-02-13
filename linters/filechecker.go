package linters

import (
	"context"
	"github.com/apple/pkl-go/pkl"
	"github.com/spf13/afero"
	"strings"
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
		return "", ErrFileInvalid.WithFile(path)
	}

	fileContent := string(bytez)
	if strings.HasSuffix(path, ".pkl") {
		evaluator, err := pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
		if err != nil {
			return content, err
		}
		defer func(evaluator pkl.Evaluator) {
			_ = evaluator.Close()
		}(evaluator)
		println(fileContent)
		return evaluator.EvaluateOutputText(context.Background(), pkl.TextSource(fileContent))
	}

	return fileContent, err
}
