package manifest

// upload-slos uploads the SLOs within the specified folder to the central Grafana.
type UploadSLOs struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"nam,omitempty"`
	// Optional path where SLOs can be found.
	Folder   string `json:"folder,omitempty" yaml:"folder,omitempty" jsonschema:"default=slos"`
	TaskBase `yaml:",inline"`
	// todo: prevent this from being included in the docs
	// todo: why do the e2e tests for upload-slos not fail for gh actions when secretAllow is not set? It fails for concourse.
	vars Vars `skipSecretsValidator:"true"`
}

func (r UploadSLOs) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r UploadSLOs) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r UploadSLOs) SetName(name string) Task {
	r.Name = name
	return r
}

func (r UploadSLOs) MarshalYAML() (any, error) {
	panic("not implemented")
	//r.Type = "upload-slos"
	//return r, nil
}

func (r UploadSLOs) GetName() string {
	if r.Name == "" {
		return "upload-slos"
	}
	return r.Name
}

func (r UploadSLOs) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r UploadSLOs) SavesArtifactsOnFailure() bool {
	return false
}

func (r UploadSLOs) SavesArtifacts() bool {
	return false
}

func (r UploadSLOs) ReadsFromArtifacts() bool {
	return false
}

func (r UploadSLOs) SetVars(vars Vars) UploadSLOs {
	r.vars = vars
	return r
}

func (r UploadSLOs) GetVars() Vars {
	return r.vars
}
