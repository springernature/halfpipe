package linters

import (
	"testing"

	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var err1 = fmt.Errorf("error1")
var err2 = fmt.Errorf("error2")

type FakeSecretValidator struct {
}

func (FakeSecretValidator) Validate(manifest.Manifest) []error {
	return []error{
		err1,
		err2,
	}
}

func TestCallsOutToSecretValidator(t *testing.T) {
	linter := NewSecretsLinter(FakeSecretValidator{})
	lintResult := linter.Lint(manifest.Manifest{})
	assert.Equal(t, []error{err1, err2}, lintResult.Issues)
}
