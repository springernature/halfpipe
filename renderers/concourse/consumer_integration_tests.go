package concourse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func convertConsumerIntegrationTestToRunTask(task manifest.ConsumerIntegrationTest, man manifest.Manifest) manifest.Run {
	consumerGitParts := strings.Split(task.Consumer, "/")
	consumerGitURI := fmt.Sprintf("git@github.com:springernature/%s", consumerGitParts[0])
	consumerGitPath := ""
	if len(consumerGitParts) > 1 {
		consumerGitPath = consumerGitParts[1]
	}

	dockerLogin := `\docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io`
	cdcScript := shared.ConsumerIntegrationTestScript(task.Vars, config.DockerComposeCacheDirs)
	script := dockerLogin + "\n\n" + cdcScript

	providerName := task.ProviderName
	if providerName == "" {
		providerName = man.Pipeline
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(providerName))

	runTask := manifest.Run{
		Retries: task.Retries,
		Name:    task.Name,
		Script:  script,
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + config.DockerComposeImage,
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: true,
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":       consumerGitURI,
			"CONSUMER_NAME":          task.Consumer,
			"CONSUMER_PATH":          consumerGitPath,
			"CONSUMER_SCRIPT":        task.Script,
			"CONSUMER_GIT_KEY":       "((halfpipe-github.private_key))",
			"CONSUMER_HOST":          task.ConsumerHost,
			"PROVIDER_NAME":          providerName,
			"PROVIDER_HOST_KEY":      providerHostKey,
			"PROVIDER_HOST":          task.ProviderHost,
			"DOCKER_COMPOSE_SERVICE": task.DockerComposeService,
			"GCR_PRIVATE_KEY":        "((halfpipe-gcr.private_key))",
			"GIT_CLONE_OPTIONS":      task.GitCloneOptions,
		},
		Timeout: task.GetTimeout(),
	}

	for key, val := range task.Vars {
		runTask.Vars[key] = val
	}

	return runTask
}

// convert string to uppercase and replace non A-Z 0-9 with underscores
func toEnvironmentKey(s string) string {
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(strings.ToUpper(s), "_")
}
