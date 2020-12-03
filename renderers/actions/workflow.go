package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/config"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name string `yaml:"name"`
	On   On     `yaml:"on"`
	Env  Env    `yaml:"env,omitempty"`
	Jobs Jobs   `yaml:"jobs,omitempty"`
}

type On struct {
	Push               Push               `yaml:"push,omitempty"`
	RepositoryDispatch RepositoryDispatch `yaml:"repository_dispatch,omitempty"`
	Schedule           []Cron             `yaml:"schedule,omitempty"`
	WorkflowDispatch   WorkflowDispatch   `yaml:"workflow_dispatch"`
}

type WorkflowDispatch struct{}

type RepositoryDispatch struct {
	Types []string `yaml:"types,omitempty"`
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

type Jobs yaml.MapSlice

type Job struct {
	Name           string    `yaml:"name,omitempty"`
	RunsOn         string    `yaml:"runs-on,omitempty"`
	Container      Container `yaml:"container,omitempty"`
	TimeoutMinutes int       `yaml:"timeout-minutes,omitempty"`
	Steps          []Step    `yaml:"steps,omitempty"`
	Env            Env       `yaml:"env,omitempty"`
}

func (j Job) ID() string {
	re := regexp.MustCompile(`[^a-z_\-]`)
	return re.ReplaceAllString(strings.ToLower(j.Name), "_")
}

type Container struct {
	Image       string
	Credentials Credentials `yaml:"credentials,omitempty"`
}

type Credentials struct {
	Username string
	Password string
}

type Step struct {
	Name string `yaml:"name,omitempty"`
	If   string `yaml:"if,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	Run  string `yaml:"run,omitempty"`
	With With   `yaml:"with,omitempty"`
}

type With yaml.MapSlice

type Env map[string]string

func (w Workflow) asYAML() (string, error) {
	output, err := yaml.Marshal(w)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("# Generated using halfpipe cli version %s\n%s", config.Version, output), nil
}
