package linters

import "fmt"
import (
	"strings"

	"path"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/errors"
)

type LintResults []LintResult
type LintResult struct {
	Linter string
	Errors []error
}

func NewLintResult(linter string, errs []error) LintResult {
	return LintResult{
		Linter: linter,
		Errors: errs,
	}
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
		for _, err := range deduplicateErrors(lr.Errors) {
			out += fmt.Sprintf("\t* %s\n", err)
			if doc, ok := err.(errors.Documented); ok {
				out += fmt.Sprintf("\t  see: %s\n", renderDocLink(doc.DocID()))
			}
			out += fmt.Sprintf("\n")
		}
	} else {
		out += fmt.Sprintf("\t%s\n\n", `No errors \o/`)
	}
	return
}

func (lr LintResult) HasErrors() bool {
	return len(lr.Errors) != 0
}

func (lr *LintResult) AddError(err ...error) {
	lr.Errors = append(lr.Errors, err...)
}

func deduplicateErrors(errs []error) (errors []error) {
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

func renderDocLink(docID string) string {
	return path.Join(config.DocHost, "/docs/cli-errors", renderDocAnchor(docID))
}

func renderDocAnchor(docID string) string {
	if docID != "" {
		return fmt.Sprintf("#%s", normalize(docID))
	}
	return ""
}

func normalize(value string) string {
	withoutWhiteSpace := strings.Replace(strings.ToLower(value), " ", "-", -1)
	return strings.Replace(withoutWhiteSpace, ".", "", -1)

}
