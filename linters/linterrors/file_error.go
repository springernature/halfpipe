package linterrors

import "fmt"

type FileError struct {
	Path   string
	Reason string
}

func NewFileError(path string, reason string) FileError {
	return FileError{Path: path, Reason: reason}
}

func (e FileError) Error() string {
	return fmt.Sprintf("'%s' %s", e.Path, e.Reason)
}
