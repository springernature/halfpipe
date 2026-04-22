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
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
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
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
}

func (r Run) GetBuildHistory() int {
	return r.BuildHistory
}

func (r Run) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}

func (r Run) GetNotifications() Notifications {
	return r.Notifications
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

func (r Run) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
