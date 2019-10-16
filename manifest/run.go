package manifest

import (
	"fmt"
	"strings"
)

type Docker struct {
	Image    string
	Username string `yaml:"username,omitempty" secretAllowed:"true"`
	Password string `yaml:"password,omitempty" secretAllowed:"true"`
}

type Run struct {
	Type                   string
	Name                   string   `yaml:"name,omitempty"`
	ManualTrigger          bool     `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	Script                 string   `yaml:"script,omitempty"`
	Docker                 Docker   `yaml:"docker,omitempty"`
	Privileged             bool     `yaml:"privileged,omitempty"`
	Vars                   Vars     `yaml:"vars,omitempty" secretAllowed:"true"`
	SaveArtifacts          []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	RestoreArtifacts       bool     `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	SaveArtifactsOnFailure []string `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
	Retries                int      `yaml:"retries,omitempty"`
	NotifyOnSuccess        bool     `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty"`
	Timeout                string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (r Run) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r Run) SetName(name string) Task {
	r.Name = name
	return r
}

func (r Run) MarshalYAML() (interface{}, error) {
	r.Type = "run"
	return r, nil
}

func (r Run) GetName() string {
	if r.Name == "" {
		return fmt.Sprintf("run %s", strings.Replace(r.Script, "./", "", 1))
	}
	return r.Name
}

func (r Run) GetTimeout() string {
	return r.Timeout
}

func (r Run) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r Run) SavesArtifactsOnFailure() bool {
	return len(r.SaveArtifactsOnFailure) > 0
}

func (r Run) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r Run) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r Run) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r Run) GetAttempts() int {
	return 1 + r.Retries
}
