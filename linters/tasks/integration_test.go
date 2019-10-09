package tasks

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, true)
	if assert.Len(t, errors, 4) {
		linterrors.AssertMissingFieldInErrors(t, "consumer", errors)
		linterrors.AssertMissingFieldInErrors(t, "consumer_host", errors)
		linterrors.AssertMissingFieldInErrors(t, "provider_host", errors)
		linterrors.AssertMissingFieldInErrors(t, "script", errors)
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsFromPrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, false)
	if assert.Len(t, errors, 3) {
		linterrors.AssertMissingFieldInErrors(t, "consumer", errors)
		linterrors.AssertMissingFieldInErrors(t, "consumer_host", errors)
		linterrors.AssertMissingFieldInErrors(t, "script", errors)
	}
}

func TestConsumerIntegrationRetries(t *testing.T) {
	task := manifest.ConsumerIntegrationTest{}

	task.Retries = -1
	errors, _ := LintConsumerIntegrationTestTask(task, false)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintConsumerIntegrationTestTask(task, false)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}
