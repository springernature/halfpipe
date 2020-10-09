package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"gopkg.in/yaml.v2"
)

func ToString(pipeline Actions) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	if err != nil {
		return "", err
	}

	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", config.Version)
	return fmt.Sprintf("%s\n%s", versionComment, renderedPipeline), nil
}
