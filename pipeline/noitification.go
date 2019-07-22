package pipeline

import "github.com/concourse/concourse/atc"

func slackOnFailurePlan(channel string) atc.PlanConfig {
	return atc.PlanConfig{
		Put: slackResourceName,
		Params: atc.Params{
			"channel":  channel,
			"username": "Halfpipe",
			"icon_url": "https://concourse.halfpipe.io/public/images/favicon-failed.png",
			"text":     "The pipeline `$BUILD_PIPELINE_NAME` failed at `$BUILD_JOB_NAME`. <$ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>",
		},
	}
}

func slackOnSuccessPlan(channel string) atc.PlanConfig {
	return atc.PlanConfig{
		Put: slackResourceName,
		Params: atc.Params{
			"channel":  channel,
			"username": "Halfpipe",
			"icon_url": "https://concourse.halfpipe.io/public/images/favicon-succeeded.png",
			"text":     "Pipeline `$BUILD_PIPELINE_NAME`, Task `$BUILD_JOB_NAME` succeeded",
		},
	}
}
