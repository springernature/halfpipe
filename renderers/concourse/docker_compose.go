package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"sort"
	"strings"
)

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest, basePath string) atc.JobConfig {
	return p.runJob(dockerComposeToRunTask(task, man), man, true, basePath)
}

func dockerComposeToRunTask(task manifest.DockerCompose, man manifest.Manifest) manifest.Run {
	if task.Vars == nil {
		task.Vars = make(map[string]string)
	}
	task.Vars["GCR_PRIVATE_KEY"] = "((halfpipe-gcr.private_key))"
	task.Vars["HALFPIPE_CACHE_TEAM"] = man.Team

	return manifest.Run{
		Retries: task.Retries,
		Name:    task.GetName(),
		Script:  dockerComposeScript(task, man.FeatureToggles.Versioned()),
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + config.DockerComposeImage,
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged:             true,
		Vars:                   task.Vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		Timeout:                task.GetTimeout(),
	}
}

func dockerComposeScript(task manifest.DockerCompose, versioningEnabled bool) string {
	envStrings := []string{"-e GIT_REVISION"}
	for key := range task.Vars {
		if key == "GCR_PRIVATE_KEY" {
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
	if task.ComposeFile != "docker-compose.yml" {
		composeFileOption = "-f " + task.ComposeFile
	}
	envOption := strings.Join(envStrings, " ")
	volumeOption := strings.Join(cacheVolumeFlags, " ")

	composeCommand := fmt.Sprintf("docker-compose %s run %s %s %s",
		composeFileOption,
		envOption,
		volumeOption,
		task.Service,
	)

	if task.Command != "" {
		composeCommand = fmt.Sprintf("%s %s", composeCommand, task.Command)
	}

	return fmt.Sprintf(`\docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io
%s
`, composeCommand)
}
