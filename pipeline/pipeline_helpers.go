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

	// tidy up secrets a bit
	name := fmt.Sprintf("CF %s %s %s", api, task.Org, task.Space)
	name = strings.Replace(name, "((cloudfoundry.", "", -1)
	name = strings.Replace(name, "))", "", -1)
	name = strings.Replace(name, "api-", "", -1)
	name = strings.Replace(name, " org-snpaas", "", -1)
	return name
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

func GenerateArtifactsResourceName(team string, pipeline string) string {
	postfix := strings.Replace(path.Join(team, pipeline), "/", "-", -1)
	return fmt.Sprintf("artifacts-%s", postfix)
}

func GenerateArtifactsOnFailureResourceName(team string, pipeline string) string {
	postfix := strings.Replace(path.Join(team, pipeline), "/", "-", -1)
	return fmt.Sprintf("artifacts-%s-on-failure", postfix)
}
func saveArtifactOnFailurePlan(team, pipeline string) atc.PlanConfig {
	return atc.PlanConfig{
		Put:      artifactsOnFailureName,
		Resource: GenerateArtifactsOnFailureResourceName(team, pipeline),
		Params: atc.Params{
			"folder":       artifactsOutDirOnFailure,
			"version_file": path.Join(gitDir, ".git", "ref"),
			"postfix":      "failure",
		},
	}
}

func slackOnFailurePlan(channel string) atc.PlanConfig {
	return atc.PlanConfig{
		Put: slackResourceName,
		Params: atc.Params{
			"channel":  channel,
			"username": "Halfpipe",
			"icon_url": "https://concourse.halfpipe.io/public/images/favicon-failed.png",
			"text":     "The pipeline `$BUILD_PIPELINE_NAME` failed at `$BUILD_JOB_NAME`. <$ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>",
		},
	}
}

func slackOnSuccessPlan(channel string) atc.PlanConfig {
	return atc.PlanConfig{
		Put: slackResourceName,
		Params: atc.Params{
			"channel":  channel,
			"username": "Halfpipe",
			"icon_url": "https://concourse.halfpipe.io/public/images/favicon-succeeded.png",
			"text":     "Pipeline `$BUILD_PIPELINE_NAME`, Task `$BUILD_JOB_NAME` succeeded",
		},
	}
}
