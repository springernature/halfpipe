package linters

import (
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var secretsLinter = SecretsLinter{}

func TestFindSecretsPlaceholder(t *testing.T) {
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.DeployCF{
			Password: "((supersecret.password))",
		},
	}

	result := secretsLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}