package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, true)
	if assert.Len(t, errors, 4) {
		AssertContainsError(t, errors, NewErrMissingField("consumer"))
		AssertContainsError(t, errors, NewErrMissingField("consumer_host"))
		AssertContainsError(t, errors, NewErrMissingField("provider_host"))
		AssertContainsError(t, errors, NewErrMissingField("script"))
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsFromPrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, false)
	if assert.Len(t, errors, 3) {
		AssertContainsError(t, errors, NewErrMissingField("consumer"))
		AssertContainsError(t, errors, NewErrMissingField("consumer_host"))
		AssertContainsError(t, errors, NewErrMissingField("script"))
	}
}

func TestConsumerIntegrationRetries(t *testing.T) {
	task := manifest.ConsumerIntegrationTest{}

	task.Retries = -1
	errors, _ := LintConsumerIntegrationTestTask(task, false)
	AssertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors, _ = LintConsumerIntegrationTestTask(task, false)
	AssertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
}
