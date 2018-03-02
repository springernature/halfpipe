package pipeline

import (
	"fmt"
	"strings"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/parser"
	"gopkg.in/yaml.v2"
	"github.com/springernature/halfpipe/config"
)

func convertVars(vars parser.Vars) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range vars {
		out[k] = v
	}
	return out
}

func deployCFResourceName(task parser.DeployCF) string {
	return fmt.Sprintf("CF %s-%s", task.Org, task.Space)
}

func getUniqueName(name string, config *atc.Config, counter int) string {
	candidate := strings.Replace(name, "/", "__", -1) //avoid bug in atc web interface
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

func ToString(pipeline atc.Config) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	version, _ := config.GetVersion()
	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", version)
	return fmt.Sprintf("%s\n%s", versionComment, renderedPipeline), err
}
