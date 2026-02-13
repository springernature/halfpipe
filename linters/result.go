package linters

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"slices"
)

type LintResults []LintResult

func (lrs LintResults) HasWarnings() bool {
	return slices.ContainsFunc(lrs, func(lr LintResult) bool { return lr.HasWarnings() })
}

func (lrs LintResults) HasErrors() bool {
	return slices.ContainsFunc(lrs, func(lr LintResult) bool { return lr.HasErrors() })
}

func (lrs LintResults) Error() (out string) {
	if lrs.HasErrors() || lrs.HasWarnings() {
		out += "The Halfpipe Linter found issues in your project:\n"
	}
	for _, result := range lrs {
		out += result.Error()
	}
	return out
}

type LintResult struct {
	Linter  string
	DocsURL string
	Issues  []error
}

func NewLintResult(linter string, docsURL string, issues []error) LintResult {
	return LintResult{
		Linter:  linter,
		DocsURL: docsURL,
		Issues:  issues,
	}
}

func (lr *LintResult) Error() (out string) {
	if lr.HasWarnings() || lr.HasErrors() {
		out += fmt.Sprintf("\n%s <%s>\n", lr.Linter, lr.DocsURL)
		for _, err := range deduplicate(lr.Issues) {
			if isWarning(err) {
				out += color.FgYellow.Sprintf("  [WARNING] %s\n", err)
			} else {
				out += color.FgRed.Sprintf("  [ERROR] %s\n", err)
			}
		}
	}
	return out
}

func (lr *LintResult) HasErrors() bool {
	return slices.ContainsFunc(lr.Issues, func(e error) bool { return !isWarning(e) })
}

func (lr *LintResult) HasWarnings() bool {
	return slices.ContainsFunc(lr.Issues, func(e error) bool { return isWarning(e) })
}

func isWarning(e error) bool {
	var lintError Error
	if ok := errors.As(e, &lintError); ok {
		return lintError.IsWarning()
	}
	return false
}

func (lr *LintResult) Add(errs ...error) {
	lr.Issues = append(lr.Issues, errs...)
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
	return slices.Contains(errs, err)
}
