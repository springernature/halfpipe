package errors

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
)

type MissingHalfpipeFileError struct {
}

func NewMissingHalfpipeFileError() MissingHalfpipeFileError {
	return MissingHalfpipeFileError{}
}

func (e MissingHalfpipeFileError) Error() string {
	return fmt.Sprintf("couldn't find any of the allowed %s files", config.HalfpipeFilenameOptions)
}
