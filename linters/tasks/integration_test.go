package tasks

import (
	"testing"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe/helpers"
)

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, "taskId", true)
	if assert.Len(t, errors, 4) {
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.consumer", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.consumer_host", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.provider_host", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.script", errors)
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsFromPrePromote(t *testing.T) {

	task := manifest.ConsumerIntegrationTest{}

	errors, _ := LintConsumerIntegrationTestTask(task, "taskId", false)
	if assert.Len(t, errors, 3) {
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.consumer", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.consumer_host", errors)
		helpers.AssertMissingFieldInErrors(t, "consumer-integration-test.script", errors)
	}
}

func TestConsumerIntegrationRetries(t *testing.T) {
	task := manifest.ConsumerIntegrationTest{}

	task.Retries = -1
	errors, _ := LintConsumerIntegrationTestTask(task, "task", false)
	helpers.AssertInvalidFieldInErrors(t, "consumer-integration-test.retries", errors)

	task.Retries = 6
	errors, _ = LintConsumerIntegrationTestTask(task, "task", false)
	helpers.AssertInvalidFieldInErrors(t, "consumer-integration-test.retries", errors)
}