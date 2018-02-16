package errors

import "fmt"

type LintResults []LintResult

func (e LintResults) HasErrors() bool {
	for _, lintResult := range e {
		if lintResult.HasErrors() {
			return true
		}
	}
	return false
}

type LintResult struct {
	Linter string
	Errors []error
}

func (lr LintResult) Error() (out string) {
	out += fmt.Sprintf("%s\n", lr.Linter)
	if lr.HasErrors() {
		for _, err := range lr.Errors {
			out += fmt.Sprintf("\t%s\n", err)
		}
	} else {
		out += fmt.Sprintf("\t%s\n", `No errors \o/`)
	}
	return
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr *LintResult) AddError(err ...error) {
	for _, e := range err {
		lr.Errors = append(lr.Errors, e)
	}
}

type InvalidField struct {
	Name   string
	Reason string
}

type Documented interface {
	DocumentationPath() string
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

type ParseError struct {
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Error parsing manifest: %s", e.Message)
}

func (e ParseError) DocumentationPath() string {
	return "/docs/manifest"
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

func (e FileError) DocumentationPath() string {
	return "/docs/manifest/required-files"
}

func NewFileError(path string, reason string) FileError {
	return FileError{Path: path, Reason: reason}
}
