package model

import "fmt"

type InvalidField struct {
	Name   string
	Reason string
}

func (e InvalidField) Error() string {
	return fmt.Sprintf("Invalid value for '%s': %s", e.Name, e.Reason)
}

func (e InvalidField) DocumentationPath() string {
	return "/docs/manifest/fields#" + e.Name
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

func (e MissingField) DocumentationPath() string {
	return "/docs/manifest/fields#" + e.Name
}

func NewMissingField(name string) MissingField {
	return MissingField{name}
}

type parseError struct {
	Message string
}

func (e parseError) Error() string {
	return fmt.Sprintf("Error parsing manifest: %s", e.Message)
}

func (e parseError) DocumentationPath() string {
	return "/docs/manifest"
}

func NewParseError(message string) parseError {
	return parseError{message}
}
