package linters

import "github.com/springernature/halfpipe/manifest"

const notificationReasons = "please migrate to new notification structure"

func LintNotifications(task manifest.Task) (errs []error) {
	switch task := task.(type) {
	case manifest.Parallel, manifest.Sequence:
		break
	default:
		n := task.GetNotifications()
		if len(n.OnSuccess) > 0 {
			errs = append(errs, NewDeprecatedField("on_success", notificationReasons).AsWarning())
		}
		if n.OnSuccessMessage != "" {
			errs = append(errs, NewDeprecatedField("on_success_message", notificationReasons).AsWarning())
		}
		if len(n.OnFailure) > 0 {
			errs = append(errs, NewDeprecatedField("on_failure", notificationReasons).AsWarning())

		}
		if len(n.OnFailureMessage) > 0 {
			errs = append(errs, NewDeprecatedField("on_failure_message", notificationReasons).AsWarning())
		}
	}

	return
}
