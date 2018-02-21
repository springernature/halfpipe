package model

import "fmt"
import (
	"strings"

	"net/url"
	"path"

	"github.com/springernature/halfpipe/errors"
)

// This field will be populated in Concourse
// go build -ldflags "-X ..."
// TODO: better env var?
var docHost = ""

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
		for _, err := range lr.Errors {
			out += fmt.Sprintf("\t* %s\n", err)
			if doc, ok := err.(errors.Documented); ok {
				out += fmt.Sprintf("\t  [see: %s ]", renderDocLink(doc.DocId()))
			}
			out += fmt.Sprintf("\n\n")
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

func renderDocLink(docId string) string {
	u, _ := url.Parse(docHost)

	return path.Join(u.Path, "/docs/linter-errors", renderDocAnchor(docId))
}

func renderDocAnchor(docId string) string {
	if docId != "" {
		return fmt.Sprintf("#%s", normalize(docId))
	}
	return ""
}

func normalize(value string) string {
	remove_white_space := strings.Replace(strings.ToLower(value), " ", "-", -1)
	return strings.Replace(remove_white_space, ".", "", -1)

}
