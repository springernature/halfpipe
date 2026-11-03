package manifest

// observability-provisioning upload o11y resources
type ObservabilityProvisioning struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"nam,omitempty"`
	// Optional path where SLOs can be found. If not set, SLOs will not be synced.
	SLOs string `json:"slos,omitempty" yaml:"slos,omitempty""`
	// Optional path where alerting and recording rules can be found. If not set, those rules will not be synced.
	MetricRules string `json:"metric_rules,omitempty" yaml:"metric_rules,omitempty""`
	TaskBase    `yaml:",inline"`
	vars        Vars `skipSecretsValidator:"true"`
}

func (r ObservabilityProvisioning) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r ObservabilityProvisioning) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r ObservabilityProvisioning) SetName(name string) Task {
	r.Name = name
	return r
}

func (r ObservabilityProvisioning) MarshalYAML() (any, error) {
	panic("not implemented")
}

func (r ObservabilityProvisioning) GetName() string {
	if r.Name == "" {
		return "observability-provisioning"
	}
	return r.Name
}

func (r ObservabilityProvisioning) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r ObservabilityProvisioning) SavesArtifactsOnFailure() bool {
	return false
}

func (r ObservabilityProvisioning) SavesArtifacts() bool {
	return false
}

func (r ObservabilityProvisioning) ReadsFromArtifacts() bool {
	return false
}

func (r ObservabilityProvisioning) SetVars(vars Vars) ObservabilityProvisioning {
	r.vars = vars
	return r
}

func (r ObservabilityProvisioning) GetVars() Vars {
	return r.vars
}
