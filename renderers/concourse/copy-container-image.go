package concourse

import (
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func convertCopyContainerImageToRunTask(task manifest.CopyContainerImage, man manifest.Manifest) manifest.Run {
	script := []string{
		"\\mkdir -p ~/.docker",
		"echo $DOCKER_CONFIG_JSON > ~/.docker/config.json\n%s",
		shared.CopyContainerImageScript,
	}

	return manifest.Run{
		Retries: task.Retries,
		Name:    task.GetName(),
		Script:  strings.Join(script, "\n"),
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + config.DockerComposeImage,
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: true,
		Vars: manifest.Vars{
			"SOURCE_URL":            task.Source,
			"TARGET_URL":            task.Target,
			"AWS_ACCESS_KEY_ID":     task.AwsAccessKeyID,
			"AWS_SECRET_ACCESS_KEY": task.AwsSecretAccessKey,
			"DOCKER_CONFIG_JSON":    "((halfpipe-gcr.docker_config))",
		},
		Timeout: task.GetTimeout(),
	}
}
