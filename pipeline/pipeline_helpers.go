package pipeline

import (
	"fmt"
	"strings"

	"regexp"

	"path"

	"github.com/concourse/concourse/atc"
	"github.com/simonjohansson/yaml"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func convertVars(vars manifest.Vars) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range vars {
		out[k] = v
	}
	return out
}

func deployCFResourceName(task manifest.DeployCF) (name string) {
	// if url remove the scheme
	api := strings.Replace(task.API, "https://", "", -1)
	api = strings.Replace(api, "http://", "", -1)
	api = strings.Replace(api, "((cloudfoundry.api-", "", -1)
	api = strings.Replace(api, "))", "", -1)
	name = fmt.Sprintf("CF %s", api)

	if org := strings.Replace(task.Org, "((cloudfoundry.org-snpaas))", "", -1); org != "" {
		name = fmt.Sprintf("%s %s", name, org)
	}

	name = fmt.Sprintf(fmt.Sprintf("%s %s", name, task.Space))
	name = strings.TrimSpace(name)
	return
}

func dockerResourceName(imageName string) string {
	parts := strings.Split(imageName, "/")
	return parts[len(parts)-1]
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

func saveArtifactOnFailurePlan() atc.PlanConfig {
	return atc.PlanConfig{
		Put: artifactsOnFailureName,
		Params: atc.Params{
			"folder":       artifactsOutDirOnFailure,
			"version_file": path.Join(gitDir, ".git", "ref"),
			"postfix":      "failure",
		},
		Attempts: artifactsPutAttempts,
	}
}
