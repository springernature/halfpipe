package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"testing"
)

func TestBuildpackTask(t *testing.T) {

	t.Run("no options set", func(t *testing.T) {
		errors := LintBuildpackTask(manifest.Buildpack{})
		assertContainsError(t, errors, NewErrMissingField("image"))
		assertContainsError(t, errors, NewErrMissingField("buildpacks"))
	})

}
