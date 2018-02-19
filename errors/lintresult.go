package errors

import "fmt"
import (
	"net/http"

	"strings"

	"github.com/asaskevich/govalidator"
)

var docBaseUrl = "https://half-pipe-landing.apps.public.gcp.springernature.io"

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
	out += fmt.Sprintf("%s %s\n", lr.Linter, helpLink(lr.Linter))
	if lr.HasErrors() {
		for _, err := range lr.Errors {
			out += fmt.Sprintf("\t%s\n", err)
		}
	} else {
		out += fmt.Sprintf("\t%s\n", `No errors \o/`)
	}
	return
}

func helpLink(linterName string) string {
	docUrl := fmt.Sprintf("%s/docs/linter/%s/errors", docBaseUrl,
		govalidator.CamelCaseToUnderscore(strings.Replace(linterName, " ", "", -1)))

	if isDocValid(docUrl) {
		return fmt.Sprintf("[see: %s]", docUrl)
	}
	return ""
}

func isDocValid(docUrl string) bool {
	resp, err := http.Head(docUrl)
	if err != nil {
		return false
	}
	//return resp.StatusCode >= 200 && resp.StatusCode < 400
	return resp.StatusCode != 0
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr *LintResult) AddError(err ...error) {
	for _, e := range err {
		lr.Errors = append(lr.Errors, e)
	}
}
