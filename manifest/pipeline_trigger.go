package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

// pipeline trigger runs the pipeline when another pipeline job has completed.
// Note that you cannot trigger on pipelines from another team.
type PipelineTrigger struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Concourse URL. Defaults to the current Concourse instance.
	ConcourseURL string `json:"concourse_url,omitempty" yaml:"concourse_url,omitempty" secretAllowed:"true"`
	// Concourse username.
	Username string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	// Concourse password.
	Password string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	// Team that owns the pipeline to trigger from. Must be the same team.
	Team string `json:"team,omitempty" yaml:"team,omitempty"`
	// Name of the pipeline to trigger from.
	Pipeline string `json:"pipeline,omitempty" yaml:"pipeline,omitempty" jsonschema:"required"`
	// Job name within the pipeline to trigger from.
	Job string `json:"job,omitempty" yaml:"job,omitempty" jsonschema:"required"`
	// Job status to trigger on.
	Status string `json:"status,omitempty" yaml:"status,omitempty" jsonschema:"default=succeeded,enum=succeeded,enum=failed,enum=errored,enum=aborted"`
}

func (p PipelineTrigger) GetTriggerAttempts() int {
	return 2
}

func (p PipelineTrigger) MarshalYAML() (any, error) {
	p.Type = "pipeline"
	return p, nil
}

func (p PipelineTrigger) GetTriggerName() string {
	name := fmt.Sprintf("%s.%s", strings.ToLower(p.Pipeline), strings.ToLower(p.Job))
	replaceSpecialChars := strings.TrimSpace(regexp.MustCompile("[^a-z0-9-.]+").ReplaceAllString(name, " "))
	return strings.Replace(replaceSpecialChars, " ", "-", -1)
}
