package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name string
}

func (w Workflow) asYAML() (string, error) {
	output, err := yaml.Marshal(w)
	if err != nil {
		return "", err
	}

	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", config.Version)
	return fmt.Sprintf("%s\n%s", versionComment, output), nil
}
