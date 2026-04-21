package concourse

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func convertCopyContainerImageToRunTask(task manifest.CopyContainerImage) manifest.Run {
	return manifest.Run{
		Retries:    task.Retries,
		Name:       task.GetName(),
		Script:     shared.CopyContainerImageScript,
		Docker:     halfpipeDockerImage,
		Privileged: true,
		Vars: manifest.Vars{
			"SOURCE_URL":            task.Source,
			"TARGET_URL":            task.Target,
			"AWS_ACCESS_KEY_ID":     task.AwsAccessKeyID,
			"AWS_SECRET_ACCESS_KEY": task.AwsSecretAccessKey,
			"GAR_TOKEN":             secrets.GARToken,
		},
		Timeout:       task.GetTimeout(),
		Notifications: task.Notifications,
	}
}
