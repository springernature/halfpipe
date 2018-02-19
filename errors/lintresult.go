package errors

import "fmt"
import (
	"github.com/asaskevich/govalidator"
)

// This field will be populated in Concourse
// go build -ldflags "-X ..."
// TODO: better env var?
var DocBaseUrl = ""

func NewLintResult(linter string, errs []error) LintResult {
	return LintResult{
		Linter: linter,
		Errors: errs,
	}
}

type LintResults []LintResult

type LintResult struct {
	Linter string
	Errors []error
}

func (lr LintResults) HasErrors() bool {
	for _, lintResult := range lr {
		if lintResult.HasErrors() {
			return true
		}
	}
	return false
}

func (lr LintResult) Error() (out string) {
	out += fmt.Sprintf("%s\n", lr.Linter)
	if lr.HasErrors() {
		for _, err := range lr.Errors {
			docId := ""
			if doc, ok := err.(Documented); ok {
				docId = doc.DocId()
			}
			out += fmt.Sprintf("\t%s\n\t%s\n", err, renderDocLink(lr.Linter, docId))
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

func renderDocLink(linterName string, docId string) string {
	return fmt.Sprintf("[see: %s%s%s]", DocBaseUrl, renderDocPath(linterName), renderDocAnchor(docId))
}

func renderDocPath(linterName string) string {
	return fmt.Sprintf("/docs/help_%s", normalize(linterName))
}

func renderDocAnchor(docId string) string {
	if docId != "" {
		return fmt.Sprintf("#%s", normalize(docId))
	}
	return ""
}

func normalize(value string) string {
	return govalidator.CamelCaseToUnderscore(govalidator.WhiteList(value, "A-Za-z_"))
}
