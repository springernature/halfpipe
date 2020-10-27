package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name string `yaml:"name,omitempty"`
	On   On     `yaml:"on,omitempty"`
}

type On struct {
	Push     Push   `yaml:"push,omitempty"`
	Schedule []Cron `yaml:"schedule,omitempty"`
}

type Cron struct {
	Expression string `yaml:"cron,omitempty"`
}

type Push struct {
	Branches Branches `yaml:"branches,omitempty"`
	Paths    Paths    `yaml:"paths,omitempty"`
}

type Branches []string
type Paths []string

func (w Workflow) asYAML() (string, error) {
	output, err := yaml.Marshal(w)
	if err != nil {
		return "", err
	}

	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", config.Version)
	return fmt.Sprintf("%s\n%s", versionComment, output), nil
}
