package linter

import (
	"testing"

	. "github.com/robwhitby/halfpipe-cli/model"
	"github.com/stretchr/testify/assert"
)

var man = Manifest{
	Team: "((secret.team))",
	Repo: Repo{PrivateKey: "((repo/key))"},
	Tasks: []Task{
		Run{Script: "((SECRET-SCRIPT))"},
	},
}

func TestFindsAllSecrets(t *testing.T) {
	expected := []string{
		"secret.team",
		"repo/key",
		"SECRET-SCRIPT",
	}
	secrets := requiredSecrets(man)
	assert.Equal(t, expected, secrets)
}

func TestFindsKeysNotInStore(t *testing.T) {
	secretChecker := func(s string) bool {
		return s != "repo/key"
	}

	missingSecrets := LintSecrets(man, secretChecker)
	expected := []error{NewMissingSecret("repo/key")}

	assert.Equal(t, expected, missingSecrets)
}
