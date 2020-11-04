package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

type Workflow struct {
	Name string `yaml:"name"`
	On   On     `yaml:"on"`
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
	Name           string `yaml:"name,omitempty"`
	RunsOn         string `yaml:"runs-on,omitempty"`
	TimeoutMinutes int    `yaml:"timeout-minutes,omitempty"`
	Steps          []Step `yaml:"steps,omitempty"`
}

func (j Job) ID() string {
	re := regexp.MustCompile(`[^a-z_\-]`)
	return re.ReplaceAllString(strings.ToLower(j.Name), "_")
}

type Step struct {
	Name string `yaml:"name,omitempty"`
	If   string `yaml:"if,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	Run  string `yaml:"run,omitempty"`
	With With   `yaml:"with,omitempty"`
	Env  Env    `yaml:"env,omitempty"`
}

type With yaml.MapSlice

type Env yaml.MapSlice

func (w Workflow) asYAML() (string, error) {
	output, err := yaml.Marshal(w)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("# Generated using halfpipe cli version %s\n%s", config.Version, output), nil
}
