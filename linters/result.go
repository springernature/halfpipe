package linters

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
)

type LintResults []LintResult

func (lrs LintResults) HasWarnings() bool {
	for _, lintResult := range lrs {
		if lintResult.HasWarnings() {
			return true
		}
	}
	return false
}

func (lrs LintResults) HasErrors() bool {
	for _, lintResult := range lrs {
		if lintResult.HasErrors() {
			return true
		}
	}
	return false
}

func (lrs LintResults) Error() (out string) {
	if lrs.HasErrors() || lrs.HasWarnings() {
		out += "The Halfpipe Linter found problems in your project:\n"
	}
	for _, result := range lrs {
		out += result.Error()
	}
	return out
}

type LintResult struct {
	Linter   string
	DocsURL  string
	Errors   []error
	Warnings []error
}

func NewLintResult(linter string, docsURL string, errs []error) LintResult {
	return LintResult{
		Linter:  linter,
		DocsURL: docsURL,
		Errors:  errs,
	}
}

func (lr LintResult) Error() (out string) {
	if lr.HasWarnings() || lr.HasErrors() {
		out += fmt.Sprintf("\n%s <%s>\n", lr.Linter, lr.DocsURL)
		out += formatErrors("ERROR", lr.Errors, color.FgRed)
		out += formatErrors("WARNING", lr.Warnings, color.FgYellow)
	}
	return out
}

func formatErrors(typeOfError string, errs []error, color color.Color) (out string) {
	for _, err := range deduplicate(errs) {
		out += color.Sprintf("  [%s] %s\n", typeOfError, err)
	}
	return out
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr LintResult) HasWarnings() bool {
	return len(lr.Warnings) != 0
}

func (lr *LintResult) Add(errs ...error) {
	for _, err := range errs {
		var lintError Error
		var isWarning bool
		if ok := errors.As(err, &lintError); ok {
			isWarning = lintError.IsWarning()
		}
		if isWarning {
			lr.Warnings = append(lr.Warnings, err)
		} else {
			lr.Errors = append(lr.Errors, err)
		}
	}
}

func deduplicate(errs []error) (errors []error) {
	for _, err := range errs {
		if !errorInErrors(err, errors) {
			errors = append(errors, err)
		}
	}
	return errors
}

func errorInErrors(err error, errs []error) bool {
	for _, e := range errs {
		if err == e {
			return true
		}
	}
	return false
}
