package linters

import "fmt"

type LintResults []LintResult
type LintResult struct {
	Linter   string
	DocsURL  string
	Errors   []error
	Warnings []error
}

func NewLintResult(linter string, docsURL string, errs []error, warns []error) LintResult {
	return LintResult{
		Linter:   linter,
		DocsURL:  docsURL,
		Errors:   errs,
		Warnings: warns,
	}
}

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
	for _, result := range lrs {
		out += result.Error()
		out += "\n"
	}
	return
}

func (lr LintResult) Error() (out string) {
	if lr.DocsURL != "" {
		out += fmt.Sprintf("%s (%s)\n", lr.Linter, lr.DocsURL)
	} else {
		out += fmt.Sprintf("%s\n", lr.Linter)
	}

	if lr.HasErrors() {
		out += formatErrors("Errors", lr.Errors)
	}
	if lr.HasWarnings() {
		out += formatErrors("Warnings", lr.Warnings)
	}

	if !lr.HasErrors() && !lr.HasWarnings() {
		out += fmt.Sprintf("\t%s\n\n", `No issues \o/`)
	}

	return
}

func formatErrors(typeOfError string, errs []error) (out string) {
	out += fmt.Sprintf("\t%s:\n", typeOfError)
	for _, err := range deduplicate(errs) {
		out += fmt.Sprintf("\t\t* %s\n", err)
	}
	return
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr LintResult) HasWarnings() bool {
	return len(lr.Warnings) != 0
}

func (lr *LintResult) AddError(err ...error) {
	lr.Errors = append(lr.Errors, err...)
}

func (lr *LintResult) AddWarning(err ...error) {
	lr.Warnings = append(lr.Warnings, err...)
}

func deduplicate(errs []error) (errors []error) {
	for _, err := range errs {
		if !errorInErrors(err, errors) {
			errors = append(errors, err)
		}
	}
	return
}

func errorInErrors(err error, errs []error) bool {
	for _, e := range errs {
		if err == e {
			return true
		}
	}
	return false
}
