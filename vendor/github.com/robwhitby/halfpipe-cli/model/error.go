package model

import (
	"fmt"
)

type Documented interface {
	DocumentationPath() string
}

type missingField struct {
	Name string
}

func (e missingField) Error() string {
	return fmt.Sprintf("Missing field: %s", e.Name)
}

func (e missingField) DocumentationPath() string {
	return "/docs/manifest/fields#" + e.Name
}

func NewMissingField(name string) missingField {
	return missingField{name}
}

type invalidField struct {
	Name   string
	Reason string
}

func (e invalidField) Error() string {
	return fmt.Sprintf("Invalid value for '%s': %s", e.Name, e.Reason)
}

func (e invalidField) DocumentationPath() string {
	return "/docs/manifest/fields#" + e.Name
}

func NewInvalidField(name string, reason string) invalidField {
	return invalidField{name, reason}
}

type fileError struct {
	Path   string
	Reason string
}

func (e fileError) Error() string {
	return fmt.Sprintf("'%s' %s", e.Path, e.Reason)
}

func (e fileError) DocumentationPath() string {
	return "/docs/manifest/required-files"
}

func NewFileError(path string, reason string) fileError {
	return fileError{Path: path, Reason: reason}
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

type missingSecret struct {
	Name string
}

func (e missingSecret) Error() string {
	return fmt.Sprintf("Secret '%s' not found", e.Name)
}

func (e missingSecret) DocumentationPath() string {
	return "/docs/manifest/secrets"
}

func NewMissingSecret(name string) missingSecret {
	return missingSecret{name}
}
