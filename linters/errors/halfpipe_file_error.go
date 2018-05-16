package errors

import "fmt"

type HalfpipeFileError struct {
	Reason string
}

func NewHalfpipeFileError(reason string) HalfpipeFileError {
	return HalfpipeFileError{Reason: reason}
}

func (e HalfpipeFileError) Error() string {
	return fmt.Sprintf("%s", e.Reason)
}
