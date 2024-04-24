package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLintNotifications(t *testing.T) {
	t.Run("does nothing if there is nothing to be done", func(t *testing.T) {
		task := manifest.Run{Notifications: manifest.Notifications{}}
		assert.Len(t, LintNotifications(task), 0)
	})

	t.Run("warns if any of the deprecated fields are used", func(t *testing.T) {
		task := manifest.DockerPush{Notifications: manifest.Notifications{
			OnSuccess:        []string{"#yo"},
			OnSuccessMessage: "blah",
			OnFailure:        []string{"#howdie"},
			OnFailureMessage: "bluh",
		}}

		result := LintNotifications(task)

		assert.Len(t, result, 4)
		assert.Contains(t, result, NewDeprecatedField("on_success", notificationReasons).AsWarning())
		assert.Contains(t, result, NewDeprecatedField("on_success_message", notificationReasons).AsWarning())
		assert.Contains(t, result, NewDeprecatedField("on_failure", notificationReasons).AsWarning())
		assert.Contains(t, result, NewDeprecatedField("on_failure_message", notificationReasons).AsWarning())
	})

	t.Run("does nothing for sequence or parallel", func(t *testing.T) {
		assert.Len(t, LintNotifications(manifest.Parallel{}), 0)
		assert.Len(t, LintNotifications(manifest.Sequence{}), 0)
	})
}
