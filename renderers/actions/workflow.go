package actions

import (
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name        string   `yaml:"name"`
	On          On       `yaml:"on"`
	Env         Env      `yaml:"env,omitempty"`
	Defaults    Defaults `yaml:"defaults,omitempty"`
	Concurrency string   `yaml:"concurrency,omitempty"`
	Jobs        Jobs     `yaml:"jobs,omitempty"`
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

type Defaults struct {
	Run Run `yaml:"run,omitempty"`
}

type Run struct {
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

type Jobs yaml.MapSlice

type Job struct {
	Name           string    `yaml:"name,omitempty"`
	Needs          []string  `yaml:"needs,omitempty"`
	RunsOn         string    `yaml:"runs-on,omitempty"`
	Container      Container `yaml:"container,omitempty"`
	TimeoutMinutes int       `yaml:"timeout-minutes,omitempty"`
	Steps          Steps     `yaml:"steps,omitempty"`
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
	Name             string `yaml:"name,omitempty"`
	If               string `yaml:"if,omitempty"`
	ID               string `yaml:"id,omitempty"`
	Uses             string `yaml:"uses,omitempty"`
	Run              string `yaml:"run,omitempty"`
	With             With   `yaml:"with,omitempty"`
	Env              Env    `yaml:"env,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

type Steps []Step

type With yaml.MapSlice

type Env map[string]string

func (w Workflow) asYAML() (string, error) {
	yaml.FutureLineWrap()
	output, err := yaml.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
