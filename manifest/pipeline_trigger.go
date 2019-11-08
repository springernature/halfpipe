package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

type PipelineTrigger struct {
	Type         string
	ConcourseURL string `yaml:"concourse_url,omitempty" secretAllowed:"true"`
	Username     string `yaml:"username,omitempty" secretAllowed:"true"`
	Password     string `yaml:"password,omitempty" secretAllowed:"true"`
	Team         string `yaml:"team,omitempty"`
	Pipeline     string `yaml:"pipeline,omitempty"`
	Job          string `yaml:"job,omitempty"`
	Status       string `yaml:"status,omitempty"`
}

func (p PipelineTrigger) GetTriggerAttempts() int {
	return 2
}

func (p PipelineTrigger) MarshalYAML() (interface{}, error) {
	p.Type = "pipeline"
	return p, nil
}

func (p PipelineTrigger) GetTriggerName() string {
	name := fmt.Sprintf("%s %s", p.Pipeline, p.Job)
	return strings.TrimSpace(regexp.MustCompile("[^a-zA-Z0-9-/]+").ReplaceAllString(name, " "))
}
