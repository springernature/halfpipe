package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func (a *Actions) consumerIntegrationTestSteps(task manifest.ConsumerIntegrationTest, man manifest.Manifest) Steps {
	runTask := convertConsumerIntegrationTestToRunTask(task, man)
	return a.runSteps(runTask)
}

func convertConsumerIntegrationTestToRunTask(task manifest.ConsumerIntegrationTest, man manifest.Manifest) manifest.Run {
	consumerGitParts := strings.Split(task.Consumer, "/")
	consumerGitURI := fmt.Sprintf("git@github.com:springernature/%s", consumerGitParts[0])
	consumerGitPath := ""
	cdcScript := ""

	if len(consumerGitParts) > 1 {
		consumerGitPath = strings.Join(consumerGitParts[1:], "/")
	}

	providerName := task.ProviderName
	if providerName == "" {
		providerName = man.Pipeline
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(providerName))

	cdcScript = shared.ConsumerIntegrationTestScript(task.Vars, []string{})

	runTask := manifest.Run{
		Retries: task.Retries,
		Name:    task.Name,

		Script: cdcScript,
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":       consumerGitURI,
			"CONSUMER_PATH":          consumerGitPath,
			"CONSUMER_SCRIPT":        task.Script,
			"CONSUMER_GIT_KEY":       githubSecrets.GitHubPrivateKey,
			"CONSUMER_HOST":          task.ConsumerHost,
			"CONSUMER_NAME":          task.Consumer,
			"PROVIDER_NAME":          providerName,
			"PROVIDER_HOST_KEY":      providerHostKey,
			"PROVIDER_HOST":          task.ProviderHost,
			"DOCKER_COMPOSE_FILE":    task.DockerComposeFile,
			"DOCKER_COMPOSE_SERVICE": task.DockerComposeService,
			"GIT_CLONE_OPTIONS":      task.GitCloneOptions,
			"USE_COVENANT":           fmt.Sprintf("%v", task.UseCovenant),
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
