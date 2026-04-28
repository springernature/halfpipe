package manifest

// deploy-katee deploys an application to Katee.
type DeployKatee struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Environment variables available to the vela manifest.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Path to the vela manifest.
	VelaManifest string `json:"vela_manifest,omitempty" yaml:"vela_manifest,omitempty" jsonschema:"default=vela.yaml"`
	// Deprecated: no longer used - safe to delete.
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=no longer used - safe to delete"`
	// Deprecated: no longer used - safe to delete.
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=no longer used - safe to delete"`
	// Vela namespace to deploy to. Defaults to katee-<team>.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Deprecated: use max_checks and check_interval instead.
	DeploymentCheckTimeout int `json:"deployment_check_timeout,omitempty" yaml:"deployment_check_timeout,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use max_checks and check_interval instead"`
	// Seconds between each deployment status check.
	CheckInterval int `json:"check_interval,omitempty" yaml:"check_interval,omitempty" jsonschema:"default=2"`
	// Maximum number of status checks before the deployment is considered failed.
	MaxChecks int `json:"max_checks,omitempty" yaml:"max_checks,omitempty" jsonschema:"default=60"`
	// GitHub environment to associate with this deployment.
	GitHubEnvironment GitHubEnvironment `json:"github_environment" yaml:"github_environment,omitempty"`
	KateeManifest     VelaManifest      `json:"-" yaml:"-"`
	TaskBase          `yaml:",inline"`
}

func (d DeployKatee) ReadsFromArtifacts() bool {
	return false
}

func (d DeployKatee) SavesArtifacts() bool {
	return false
}

func (d DeployKatee) SavesArtifactsOnFailure() bool {
	return false
}

func (d DeployKatee) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	d.NotifyOnSuccess = notifyOnSuccess
	return d
}

func (d DeployKatee) SetTimeout(timeout string) Task {
	d.Timeout = timeout
	return d
}

func (d DeployKatee) GetName() string {
	if d.Name == "" {
		return "deploy-katee"
	}
	return d.Name
}

func (d DeployKatee) SetName(name string) Task {
	d.Name = name
	return d
}

func (d DeployKatee) SetNotifications(notifications Notifications) Task {
	d.Notifications = notifications
	return d
}

func (d DeployKatee) MarshalYAML() (any, error) {
	d.Type = "deploy-katee"
	return d, nil
}

func (d DeployKatee) GetGitHubEnvironment() GitHubEnvironment {
	return d.GitHubEnvironment
}
