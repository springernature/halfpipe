package manifest

// deploy-ml-modules deploys a version of the shared ml modules library from
// artifactory.
type DeployMLModules struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Version of the ml-modules artifact in Artifactory.
	MLModulesVersion string `json:"ml_modules_version" yaml:"ml_modules_version,omitempty" jsonschema:"required"`
	// MarkLogic instances to deploy to.
	Targets []string `json:"targets,omitempty" yaml:"targets,omitempty" secretAllowed:"true" jsonschema:"required"`
	// App name in MarkLogic. Defaults to the pipeline name.
	AppName string `json:"app_name" yaml:"app_name,omitempty"`
	// App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version.
	AppVersion string `json:"app_version" yaml:"app_version,omitempty"`
	// Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version.
	UseBuildVersion bool `json:"use_build_version,omitempty" yaml:"use_build_version,omitempty"`
	// Username to connect to MarkLogic. Defaults to the shared vault secret.
	Username string `json:"username" yaml:"username,omitempty" secretAllowed:"true"`
	// Password to connect to MarkLogic. Defaults to the shared vault secret.
	Password string `json:"password" yaml:"password,omitempty" secretAllowed:"true"`
	TaskBase `yaml:",inline"`
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

func (r DeployMLModules) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DeployMLModules) SavesArtifactsOnFailure() bool {
	return false
}

func (r DeployMLModules) SavesArtifacts() bool {
	return false
}

func (r DeployMLModules) ReadsFromArtifacts() bool {
	return false
}
