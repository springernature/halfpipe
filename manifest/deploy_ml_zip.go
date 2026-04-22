package manifest

// DeployMLZip deploys local XQuery files to MarkLogic using ml-deploy.
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
	// Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version.
	UseBuildVersion bool `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
	// Username to connect to MarkLogic. Defaults to the shared vault secret.
	Username string `json:"username" yaml:"username,omitempty" secretAllowed:"true"`
	// Password to connect to MarkLogic. Defaults to the shared vault secret.
	Password string `json:"password" yaml:"password,omitempty" secretAllowed:"true"`
	TaskBase `yaml:",inline"`
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

func (r DeployMLZip) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DeployMLZip) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLZip) SavesArtifacts() bool {
	return false
}

func (r DeployMLZip) ReadsFromArtifacts() bool {
	return true
}
