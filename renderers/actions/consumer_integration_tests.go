package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
	"regexp"
	"strings"
)

func (a Actions) consumerIntegrationTestJob(task manifest.ConsumerIntegrationTest, man manifest.Manifest) Job {
	runTask := convertConsumerIntegrationTestToRunTask(task, man)
	return a.runJob(runTask)
}

func convertConsumerIntegrationTestToRunTask(task manifest.ConsumerIntegrationTest, man manifest.Manifest) manifest.Run {
	consumerGitParts := strings.Split(task.Consumer, "/")
	consumerGitURI := fmt.Sprintf("git@github.com:springernature/%s", consumerGitParts[0])
	consumerGitPath := ""
	if len(consumerGitParts) > 1 {
		consumerGitPath = consumerGitParts[1]
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(man.Pipeline))

	runTask := manifest.Run{
		Retries: task.Retries,
		Name:    task.Name,
		Script:  shared.ConsumerIntegrationTestScript(task.Vars, []string{}),
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":       consumerGitURI,
			"CONSUMER_PATH":          consumerGitPath,
			"CONSUMER_SCRIPT":        task.Script,
			"CONSUMER_GIT_KEY":       "${{ secrets.EE_GITHUB_PRIVATE_KEY }}",
			"CONSUMER_HOST":          task.ConsumerHost,
			"PROVIDER_NAME":          man.Pipeline,
			"PROVIDER_HOST_KEY":      providerHostKey,
			"PROVIDER_HOST":          task.ProviderHost,
			"DOCKER_COMPOSE_SERVICE": task.DockerComposeService,
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
