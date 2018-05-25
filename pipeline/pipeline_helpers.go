package pipeline

import (
	"fmt"
	"strings"

	"regexp"

	"path"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
)

func convertVars(vars manifest.Vars) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range vars {
		out[k] = v
	}
	return out
}

func deployCFResourceName(task manifest.DeployCF) string {
	// if url remove the scheme
	api := strings.Replace(task.API, "https://", "", -1)
	api = strings.Replace(api, "http://", "", -1)

	// if secret, tidy up
	api = strings.Replace(api, "((cloudfoundry.", "", -1)
	api = strings.Replace(api, "))", "", -1)
	return fmt.Sprintf("CF %s %s %s", api, task.Org, task.Space)
}

func uniqueName(cfg *atc.Config, name string, defaultName string) string {
	if name == "" {
		name = defaultName
	}
	return getUniqueName(name, cfg, 0)
}

func getUniqueName(name string, config *atc.Config, counter int) string {
	candidate := strings.Replace(name, "/", "_", -1) //avoid bug in atc web interface
	if counter > 0 {
		candidate = fmt.Sprintf("%s (%v)", name, counter)
	}

	for _, job := range config.Jobs {
		if job.Name == candidate {
			return getUniqueName(name, config, counter+1)
		}
	}
	for _, res := range config.Resources {
		if res.Name == candidate {
			return getUniqueName(name, config, counter+1)
		}
	}
	return candidate
}

// convert string to uppercase and replace non A-Z 0-9 with underscores
func toEnvironmentKey(s string) string {
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(strings.ToUpper(s), "_")
}

func ToString(pipeline atc.Config) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	if err != nil {
		return "", err
	}

	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", config.Version)
	return fmt.Sprintf("%s\n%s", versionComment, renderedPipeline), nil
}

func GenerateArtifactsFolderName(team string, pipeline string) string {
	postfix := strings.Replace(path.Join(team, pipeline), "/", "-", -1)
	return fmt.Sprintf("artifacts-%s", postfix)
}

// convert string to be suitable for use as SerialGroup key
func toSerialGroupKey(s string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9\-]`).ReplaceAllString(s, "_")
}
