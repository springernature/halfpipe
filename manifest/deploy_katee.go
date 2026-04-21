package manifest

type DeployKatee struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Environment variables available to the vela manifest.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Path to the vela manifest. Defaults to vela.yaml.
	VelaManifest string `json:"vela_manifest,omitempty" yaml:"vela_manifest,omitempty"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Deprecated: no longer used - safe to delete.
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
	// Deprecated: no longer used - safe to delete.
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=no longer used - safe to delete"`
	// Vela namespace to deploy to. Defaults to katee-<team>.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Deprecated: use max_checks and check_interval instead.
	DeploymentCheckTimeout int `json:"deployment_check_timeout,omitempty" yaml:"deployment_check_timeout,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use max_checks and check_interval instead"`
	// Seconds between each deployment status check. Defaults to 2.
	CheckInterval int `json:"check_interval,omitempty" yaml:"check_interval,omitempty"`
	// Maximum number of status checks before the deployment is considered failed. Defaults to 60.
	MaxChecks int `json:"max_checks,omitempty" yaml:"max_checks,omitempty"`
	// GitHub environment to associate with this deployment.
	GitHubEnvironment GitHubEnvironment `json:"github_environment" yaml:"github_environment,omitempty"`
	KateeManifest     VelaManifest      `json:"-" yaml:"-"`
}

func (d DeployKatee) ReadsFromArtifacts() bool {
	return false
}

func (d DeployKatee) GetAttempts() int {
	return 2 + d.Retries
}

func (d DeployKatee) SavesArtifacts() bool {
	return false
}

func (d DeployKatee) SavesArtifactsOnFailure() bool {
	return false
}

func (d DeployKatee) IsManualTrigger() bool {
	return d.ManualTrigger
}

func (d DeployKatee) NotifiesOnSuccess() bool {
	return d.NotifyOnSuccess
}
func (r DeployKatee) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}
func (d DeployKatee) GetTimeout() string {
	return d.Timeout
}

func (d DeployKatee) SetTimeout(timeout string) Task {
	d.Timeout = timeout
	return d
}

func (r DeployKatee) GetName() string {
	if r.Name == "" {
		return "deploy-katee"
	}
	return r.Name
}

func (r DeployKatee) SetName(name string) Task {
	r.Name = name
	return r
}

func (d DeployKatee) GetNotifications() Notifications {
	return d.Notifications
}

func (d DeployKatee) SetNotifications(notifications Notifications) Task {
	d.Notifications = notifications
	return d
}

func (d DeployKatee) GetBuildHistory() int {
	return d.BuildHistory
}

func (d DeployKatee) SetBuildHistory(buildHistory int) Task {
	d.BuildHistory = buildHistory
	return d
}

func (d DeployKatee) MarshalYAML() (any, error) {
	d.Type = "deploy-katee"
	return d, nil
}

func (d DeployKatee) GetGitHubEnvironment() GitHubEnvironment {
	return d.GitHubEnvironment
}
