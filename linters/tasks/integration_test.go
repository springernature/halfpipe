package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, true)
	if assert.Len(t, errors, 4) {
		helpers.AssertMissingFieldInErrors(t, "consumer", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer_host", errors)
		helpers.AssertMissingFieldInErrors(t, "provider_host", errors)
		helpers.AssertMissingFieldInErrors(t, "script", errors)
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsFromPrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, false)
	if assert.Len(t, errors, 3) {
		helpers.AssertMissingFieldInErrors(t, "consumer", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer_host", errors)
		helpers.AssertMissingFieldInErrors(t, "script", errors)
	}
}

func TestConsumerIntegrationRetries(t *testing.T) {
	task := manifest.ConsumerIntegrationTest{}

	task.Retries = -1
	errors, _ := LintConsumerIntegrationTestTask(task, false)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintConsumerIntegrationTestTask(task, false)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)
}
