package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
)

func teamsOnSuccessPlan(webhook string, message string) atc.Step {
	return teamsPlan(webhook, "succeeded", message, "✅ it passed", "28a745")
}

func teamsOnFailurePlan(webhook string, message string) atc.Step {
	return teamsPlan(webhook, "failed", message, "❌ the bloomin thing failed", "dc3545")
}

func teamsPlan(webhook string, status string, message string, title string, color string) atc.Step {
	icon := fmt.Sprintf("https://concourse.halfpipe.io/public/images/favicon-%s.png", status)

	text := message
	if text == "" {
		text = fmt.Sprintf("Pipeline `$BUILD_PIPELINE_NAME` task `$BUILD_JOB_NAME` %s.", status)
	}

	return atc.Step{
		Config: &atc.PutStep{
			Name: teamsResourceName,
			Params: atc.Params{
				"webhook": webhook,
				"text":    text,
				"title":   title,
				"icon":    icon,
			},
			NoGet: true,
		},
	}
}
