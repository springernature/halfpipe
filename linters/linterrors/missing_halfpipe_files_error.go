package linterrors

import (
	"fmt"
)

type MissingHalfpipeFileError struct {
	HalfpipeFilenameOptions []string
}

func NewMissingHalfpipeFileError(halfpipeFilenameOptions []string) MissingHalfpipeFileError {
	return MissingHalfpipeFileError{
		HalfpipeFilenameOptions: halfpipeFilenameOptions,
	}
}

func (e MissingHalfpipeFileError) Error() string {
	if len(e.HalfpipeFilenameOptions) == 1 {
		return fmt.Sprintf("couldn't find '%s'", e.HalfpipeFilenameOptions[0])
	}
	return fmt.Sprintf("couldn't find any of the allowed %s files", e.HalfpipeFilenameOptions)
}
