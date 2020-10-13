package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

type PipelineTrigger struct {
	Type         string
	ConcourseURL string `json:"concourse_url,omitempty" yaml:"concourse_url,omitempty" secretAllowed:"true"`
	Username     string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	Password     string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	Team         string `json:"team,omitempty" yaml:"team,omitempty"`
	Pipeline     string `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	Job          string `json:"job,omitempty" yaml:"job,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
}

func (p PipelineTrigger) GetTriggerAttempts() int {
	return 2
}

func (p PipelineTrigger) MarshalYAML() (interface{}, error) {
	p.Type = "pipeline"
	return p, nil
}

func (p PipelineTrigger) GetTriggerName() string {
	name := fmt.Sprintf("%s.%s", strings.ToLower(p.Pipeline), strings.ToLower(p.Job))
	replaceSpecialChars := strings.TrimSpace(regexp.MustCompile("[^a-z0-9-.]+").ReplaceAllString(name, " "))
	return strings.Replace(replaceSpecialChars, " ", "-", -1)
}
