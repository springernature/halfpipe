package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
)

func TestCopyContainerImageTask(t *testing.T) {

	t.Run("missing fields", func(t *testing.T) {
		errors := LintCopyContainerImageTask(manifest.CopyContainerImage{})
		assertContainsError(t, errors, NewErrMissingField("source"))
		assertContainsError(t, errors, NewErrMissingField("target"))
	})

	t.Run("source should be halfpipe repo", func(t *testing.T) {
		bad := LintCopyContainerImageTask(manifest.CopyContainerImage{
			Source: "eu.gcr.io/foo/bar",
		})
		assertContainsError(t, bad, ErrCopyContainerSource)

		good := LintCopyContainerImageTask(manifest.CopyContainerImage{
			Source: "eu.gcr.io/halfpipe-io/",
		})
		assertNotContainsError(t, good, ErrCopyContainerSource)
	})
}
