package manifest

import (
	"fmt"
	"strings"
)

type Docker struct {
	// Path to docker image
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// Username for private Docker registries.
	Username string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	// Password for private Docker registries.
	Password   string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	Entrypoint string `json:"-" yaml:"-"`
}

// Run is the most generic piece of work you can do. It represents a job in a
// pipeline where a script will be run in a docker container. If the script
// returns a non-zero exit code the task will be considered failed and any
// subsequent tasks will not run.
type Run struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Path to the script to execute, relative to the manifest. Prefix with \ to run a system command e.g. \make.
	Script string `json:"script,omitempty" yaml:"script,omitempty"`
	// Docker configuration for the task to run in.
	Docker Docker `json:"docker" yaml:"docker,omitempty"`
	// Run the task as root. Not recommended but sometimes necessary e.g. for docker-in-docker.
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	// Environment variables available to the script.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Paths to files or directories to save for use in subsequent tasks.
	SaveArtifacts []string `json:"save_artifacts" yaml:"save_artifacts,omitempty"`
	// Restore artifacts saved by previous tasks.
	RestoreArtifacts bool `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	// Paths to save when the task fails, useful for test reports.
	SaveArtifactsOnFailure []string `json:"save_artifacts_on_failure" yaml:"save_artifacts_on_failure,omitempty"`
	TaskBase               `yaml:",inline"`
}

func (r Run) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r Run) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r Run) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r Run) SetName(name string) Task {
	r.Name = name
	return r
}

func (r Run) MarshalYAML() (any, error) {
	r.Type = "run"
	return r, nil
}

func (r Run) GetName() string {
	if r.Name == "" {
		return fmt.Sprintf("run %s", strings.Replace(r.Script, "./", "", 1))
	}
	return r.Name
}

func (r Run) SavesArtifactsOnFailure() bool {
	return len(r.SaveArtifactsOnFailure) > 0
}

func (r Run) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r Run) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r Run) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
