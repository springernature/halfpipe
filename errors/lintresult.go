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

func NewLintResult(linter string, errs []error) LintResult {
	return LintResult{
		Linter: linter,
		Errors: errs,
	}
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
