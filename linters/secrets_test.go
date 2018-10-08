package linters

import (
	"testing"

	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var e1 = fmt.Errorf("error1")
var e2 = fmt.Errorf("error2")

type FakeSecretValidator struct {
}

func (FakeSecretValidator) Validate(manifest.Manifest) []error {
	return []error{
		e1,
		e2,
	}
}

func TestCallsOutToSecretValidator(t *testing.T) {
	linter := NewSecretsLinter(FakeSecretValidator{})
	lintResult := linter.Lint(manifest.Manifest{})
	assert.Equal(t, []error{e1, e2}, lintResult.Errors)
}
