package errors

import "fmt"

type InvalidField struct {
	Name   string
	Reason string
}

func (e InvalidField) Error() string {
	return fmt.Sprintf("Invalid value for '%s': %s", e.Name, e.Reason)
}

func NewInvalidField(name string, reason string) InvalidField {
	return InvalidField{name, reason}
}

type MissingField struct {
	Name string
}

func (e MissingField) Error() string {
	return fmt.Sprintf("Missing field: %s", e.Name)
}

func NewMissingField(name string) MissingField {
	return MissingField{name}
}

type ParseError struct {
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Error parsing manifest: %s", e.Message)
}

func NewParseError(message string) ParseError {
	return ParseError{message}
}

type FileError struct {
	Path   string
	Reason string
}

func (e FileError) Error() string {
	return fmt.Sprintf("'%s' %s", e.Path, e.Reason)
}

func NewFileError(path string, reason string) FileError {
	return FileError{Path: path, Reason: reason}
}
