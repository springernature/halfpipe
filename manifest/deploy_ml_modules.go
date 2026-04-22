package manifest

// DeployMLModules deploys a version of the shared ml modules library from
// artifactory.
type DeployMLModules struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Version of the ml-modules artifact in Artifactory.
	MLModulesVersion string `json:"ml_modules_version" yaml:"ml_modules_version,omitempty"`
	// App name in MarkLogic. Defaults to the pipeline name.
	AppName string `json:"app_name" yaml:"app_name,omitempty"`
	// App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version.
	AppVersion string `json:"app_version" yaml:"app_version,omitempty"`
	// MarkLogic instances to deploy to.
	Targets []string `json:"targets,omitempty" yaml:"targets,omitempty" secretAllowed:"true"`
	// Task must be manually triggered (Concourse only).
	ManualTrigger bool `json:"manual_trigger" yaml:"manual_trigger,omitempty"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Timeout duration for the task. If exceeded the task fails. Defaults to 1h.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version.
	UseBuildVersion bool `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
	// Username to connect to MarkLogic. Defaults to the shared vault secret.
	Username string `json:"username" yaml:"username,omitempty" secretAllowed:"true"`
	// Password to connect to MarkLogic. Defaults to the shared vault secret.
	Password string `json:"password" yaml:"password,omitempty" secretAllowed:"true"`
	// Number of build logs to retain. Defaults to 20 (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty"`
}

func (r DeployMLModules) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DeployMLModules) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}

func (r DeployMLModules) GetNotifications() Notifications {
	return r.Notifications
}

func (r DeployMLModules) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DeployMLModules) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DeployMLModules) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DeployMLModules) MarshalYAML() (any, error) {
	r.Type = "deploy-ml-modules"
	return r, nil
}

func (r DeployMLModules) GetName() string {
	if r.Name == "" {
		return "deploy-ml-modules"
	}
	return r.Name
}

func (r DeployMLModules) GetTimeout() string {
	return r.Timeout
}

func (r DeployMLModules) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}
func (r DeployMLModules) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DeployMLModules) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLModules) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DeployMLModules) SavesArtifacts() bool {
	return false
}

func (r DeployMLModules) ReadsFromArtifacts() bool {
	return false
}

func (r DeployMLModules) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLModules) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
