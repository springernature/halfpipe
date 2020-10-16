package pipeline

import (
	"fmt"
	"github.com/concourse/concourse/atc"
)

func slackOnFailurePlan(channel string, message string) atc.Step {
	return slackPlan(channel, "failed", message)
}

func slackOnSuccessPlan(channel string, message string) atc.Step {
	return slackPlan(channel, "succeeded", message)
}

func slackPlan(channel string, status string, message string) atc.Step {
	url := "<$ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>"
	icon := fmt.Sprintf("https://concourse.halfpipe.io/public/images/favicon-%s.png", status)

	text := message
	if text == "" {
		text = fmt.Sprintf("Pipeline `$BUILD_PIPELINE_NAME`, task `$BUILD_JOB_NAME` %s %s", status, url)
	}

	return atc.Step{
		Config: &atc.PutStep{
			Name: slackResourceName,
			Params: atc.Params{
				"channel":  channel,
				"username": "Halfpipe",
				"icon_url": icon,
				"text":     text,
			},
		},
		UnknownFields: nil,
	}
}
