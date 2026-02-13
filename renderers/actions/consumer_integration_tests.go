package actions

import (
	"fmt"
	"maps"
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

	var keys []string
	for k := range task.Vars {
		keys = append(keys, k)
	}
	// In the Concourse renderer we default these to be part of task.Vars but not in Actions since
	// ARTIFACTORY_* is always using the top level env, so we can safely assume
	// that they are available for us here.
	keys = append(keys, "ARTIFACTORY_URL")
	keys = append(keys, "ARTIFACTORY_USERNAME")
	keys = append(keys, "ARTIFACTORY_PASSWORD")

	var cacheDirs = []shared.CacheDirs{
		{RunnerDir: fmt.Sprintf("/mnt/halfpipe-cache/%s", man.Team), ContainerDir: "/var/halfpipe/shared-cache"},
	}
	cdcScript = shared.ConsumerIntegrationTestScript(keys, cacheDirs, false)

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
		Timeout:                task.GetTimeout(),
		SaveArtifacts:          task.SaveArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
	}

	maps.Copy(runTask.Vars, task.Vars)

	return runTask
}

// convert string to uppercase and replace non A-Z 0-9 with underscores
func toEnvironmentKey(s string) string {
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(strings.ToUpper(s), "_")
}
