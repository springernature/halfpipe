package manifest

import (
	"encoding/json"
	"strings"
)

type DockerCompose struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Command to run against the service. If omitted the default service command is used.
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	// Environment variables available to docker-compose.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Name of the docker-compose service to run. Defaults to app.
	Service string `json:"service,omitempty" yaml:"service,omitempty"`
	// Path(s) to docker-compose file(s), space-separated. Defaults to docker-compose.yml.
	ComposeFiles ComposeFiles `json:"compose_file" yaml:"compose_file,omitempty"`
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

func (r DockerCompose) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DockerCompose) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}

func (r DockerCompose) GetNotifications() Notifications {
	return r.Notifications
}

func (r DockerCompose) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DockerCompose) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DockerCompose) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerCompose) MarshalYAML() (any, error) {
	r.Type = "docker-compose"
	return r, nil
}

func (r DockerCompose) GetName() string {
	if r.Name == "" {
		return "docker-compose"
	}
	return r.Name
}

func (r DockerCompose) GetTimeout() string {
	return r.Timeout
}

func (r DockerCompose) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}

func (r DockerCompose) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DockerCompose) SavesArtifactsOnFailure() bool {
	return len(r.SaveArtifactsOnFailure) > 0
}

func (r DockerCompose) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DockerCompose) SavesArtifacts() bool {
	return len(r.SaveArtifacts) > 0
}

func (r DockerCompose) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}

func (r DockerCompose) GetAttempts() int {
	return 1 + r.Retries
}

func (r DockerCompose) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}

type ComposeFiles []string

func (c ComposeFiles) MarshalYAML() (any, error) {
	return strings.Join(c, " "), nil
}

func (c *ComposeFiles) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	for s := range strings.SplitSeq(raw, " ") {
		*c = append(*c, s)
	}

	return nil
}
