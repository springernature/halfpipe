package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors := LintConsumerIntegrationTestTask(task, true)
	if assert.Len(t, errors, 4) {
		assertContainsError(t, errors, NewErrMissingField("consumer"))
		assertContainsError(t, errors, NewErrMissingField("consumer_host"))
		assertContainsError(t, errors, NewErrMissingField("provider_host"))
		assertContainsError(t, errors, NewErrMissingField("script"))
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsFromPrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors := LintConsumerIntegrationTestTask(task, false)
	if assert.Len(t, errors, 3) {
		assertContainsError(t, errors, NewErrMissingField("consumer"))
		assertContainsError(t, errors, NewErrMissingField("consumer_host"))
		assertContainsError(t, errors, NewErrMissingField("script"))
	}
}

func TestConsumerIntegrationRetries(t *testing.T) {
	task := manifest.ConsumerIntegrationTest{}

	task.Retries = -1
	errors := LintConsumerIntegrationTestTask(task, false)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors = LintConsumerIntegrationTestTask(task, false)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
}
