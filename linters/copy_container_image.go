package linters

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func LintCopyContainerImageTask(task manifest.CopyContainerImage) (errs []error) {

	if task.Source == "" {
		errs = append(errs, NewErrMissingField("source"))
	} else if !strings.HasPrefix(task.Source, "eu.gcr.io/halfpipe-io/") {
		errs = append(errs, ErrCopyContainerSource.WithValue(task.Source))
	}

	if len(task.Target) == 0 {
		errs = append(errs, NewErrMissingField("target"))
	}

	return errs
}
