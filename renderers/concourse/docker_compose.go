package concourse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func convertDockerComposeToRunTask(task manifest.DockerCompose, man manifest.Manifest) manifest.Run {
	if task.Vars == nil {
		task.Vars = make(map[string]string)
	}
	task.Vars["GAR_TOKEN"] = vaultSecrets.GARToken
	task.Vars["HALFPIPE_CACHE_TEAM"] = man.Team

	return manifest.Run{
		Retries:                task.Retries,
		Name:                   task.GetName(),
		Script:                 dockerComposeScript(task, man.FeatureToggles.UpdatePipeline()),
		Docker:                 halfpipeDockerImage,
		Privileged:             true,
		Vars:                   task.Vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		Timeout:                task.GetTimeout(),
	}
}

func dockerComposeScript(task manifest.DockerCompose, versioningEnabled bool) string {
	envStrings := []string{"-e GIT_REVISION", `-e DOCKER_HOST="${DIND_HOST}"`}
	for key := range task.Vars {
		if key == "GAR_TOKEN" {
			continue
		}
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	if versioningEnabled {
		envStrings = append(envStrings, "-e BUILD_VERSION")
	}
	sort.Strings(envStrings)

	var cacheVolumeFlags []string
	for _, cacheVolume := range config.DockerComposeCacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cacheVolume, cacheVolume))
	}

	composeFileOption := ""
	for _, f := range task.ComposeFiles {
		composeFileOption += fmt.Sprintf(" -f %s", f)
	}
	if composeFileOption == " -f docker-compose.yml" {
		composeFileOption = ""
	}

	envOption := strings.Join(envStrings, " ")
	volumeOption := strings.Join(cacheVolumeFlags, " ")

	composeCommand := fmt.Sprintf(`\docker-compose%s run --use-aliases %s %s %s`,
		composeFileOption,
		envOption,
		volumeOption,
		task.Service,
	)

	if task.Command != "" {
		composeCommand = fmt.Sprintf("%s %s", composeCommand, task.Command)
	}

	return fmt.Sprintf("%s\n", composeCommand)
}
