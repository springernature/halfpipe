package manifest

type DeployMLZip struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Path to the zip file containing XQuery source files, relative to the manifest.
	DeployZip string `json:"deploy_zip" yaml:"deploy_zip,omitempty"`
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

func (r DeployMLZip) GetBuildHistory() int {
	return r.BuildHistory
}

func (r DeployMLZip) SetBuildHistory(buildHistory int) Task {
	r.BuildHistory = buildHistory
	return r
}

func (r DeployMLZip) GetNotifications() Notifications {
	return r.Notifications
}

func (r DeployMLZip) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DeployMLZip) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DeployMLZip) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DeployMLZip) MarshalYAML() (any, error) {
	r.Type = "deploy-ml-zip"
	return r, nil
}

func (r DeployMLZip) GetName() string {
	if r.Name == "" {
		return "deploy-ml-zip"
	}
	return r.Name
}

func (r DeployMLZip) GetTimeout() string {
	return r.Timeout
}

func (r DeployMLZip) NotifiesOnSuccess() bool {
	return r.NotifyOnSuccess
}
func (r DeployMLZip) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}
func (r DeployMLZip) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLZip) IsManualTrigger() bool {
	return r.ManualTrigger
}

func (r DeployMLZip) SavesArtifacts() bool {
	return false
}

func (r DeployMLZip) GetAttempts() int {
	return 1 + r.Retries
}

func (r DeployMLZip) ReadsFromArtifacts() bool {
	return true
}

func (r DeployMLZip) GetGitHubEnvironment() GitHubEnvironment {
	return GitHubEnvironment{}
}
