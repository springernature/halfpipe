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
	Linter  string
	DocsURL string
	Errors  []error
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
		for _, err := range deduplicate(lr.Errors) {
			if isWarning(err) {
				out += color.FgYellow.Sprintf("  [WARNING] %s\n", err)
			} else {
				out += color.FgRed.Sprintf("  [ERROR] %s\n", err)
			}
		}
	}
	return out
}

func (lr LintResult) HasErrors() bool {
	for _, e := range lr.Errors {
		if !isWarning(e) {
			return true
		}
	}
	return false
}

func (lr LintResult) HasWarnings() bool {
	for _, e := range lr.Errors {
		if isWarning(e) {
			return true
		}
	}
	return false
}

func isWarning(e error) bool {
	var lintError Error
	if ok := errors.As(e, &lintError); ok {
		return lintError.IsWarning()
	}
	return false
}

func (lr *LintResult) Add(errs ...error) {
	lr.Errors = append(lr.Errors, errs...)
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
