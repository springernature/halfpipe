package pipeline

import (
	"fmt"
	"github.com/concourse/concourse/atc"
)

func slackOnFailurePlan(channel string) atc.PlanConfig {
	return slackPlan(channel, "failed")
}

func slackOnSuccessPlan(channel string) atc.PlanConfig {
	return slackPlan(channel, "succeeded")
}

func slackPlan(channel string, status string) atc.PlanConfig {
	url := "<$ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>"
	return atc.PlanConfig{
		Put: slackResourceName,
		Params: atc.Params{
			"channel":  channel,
			"username": "Halfpipe",
			"icon_url": fmt.Sprintf("https://concourse.halfpipe.io/public/images/favicon-%s.png", status),
			"text":     fmt.Sprintf("Pipeline `$BUILD_PIPELINE_NAME`, task `$BUILD_JOB_NAME` %s %s", status, url),
		},
	}
}
