package concourse

import (
	"fmt"
	"maps"
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
		consumerGitPath = strings.Join(consumerGitParts[1:], "/")
	}

	var keys []string
	for k := range task.Vars {
		keys = append(keys, k)
	}
	var cacheDirs []shared.CacheDirs
	for _, cacheDir := range config.DockerComposeCacheDirs {
		cacheDirs = append(cacheDirs, shared.CacheDirs{RunnerDir: cacheDir, ContainerDir: cacheDir})
	}

	providerName := task.ProviderName
	if providerName == "" {
		providerName = man.Pipeline
	}
	providerHostKey := fmt.Sprintf("%s_DEPLOYED_HOST", toEnvironmentKey(providerName))

	runTask := manifest.Run{
		Retries:    task.Retries,
		Name:       task.Name,
		Script:     shared.ConsumerIntegrationTestScript(keys, cacheDirs, true),
		Docker:     halfpipeDockerImage,
		Privileged: true,
		Vars: manifest.Vars{
			"CONSUMER_GIT_URI":       consumerGitURI,
			"CONSUMER_NAME":          task.Consumer,
			"CONSUMER_PATH":          consumerGitPath,
			"CONSUMER_SCRIPT":        task.Script,
			"CONSUMER_GIT_KEY":       vaultSecrets.GitHubPrivateKey,
			"CONSUMER_HOST":          task.ConsumerHost,
			"PROVIDER_NAME":          providerName,
			"PROVIDER_HOST_KEY":      providerHostKey,
			"PROVIDER_HOST":          task.ProviderHost,
			"DOCKER_COMPOSE_FILE":    task.DockerComposeFile,
			"DOCKER_COMPOSE_SERVICE": task.DockerComposeService,
			"GAR_TOKEN":              vaultSecrets.GARToken,
			"GIT_CLONE_OPTIONS":      task.GitCloneOptions,
			"HALFPIPE_CACHE_TEAM":    man.Team,
			"USE_COVENANT":           fmt.Sprintf("%v", task.UseCovenant),
		},
		Timeout:                task.GetTimeout(),
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		SaveArtifacts:          task.SaveArtifacts,
	}

	maps.Copy(runTask.Vars, task.Vars)
	return runTask
}

// convert string to uppercase and replace non A-Z 0-9 with underscores
func toEnvironmentKey(s string) string {
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(strings.ToUpper(s), "_")
}
