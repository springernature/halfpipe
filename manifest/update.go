package manifest

type Update struct {
	Type          string
	Notifications Notifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Timeout       string
	BuildHistory  int
	TagRepo       bool
}

func (u Update) GetSecrets() map[string]string {
	panic("GetSecrets should never be called on Update")
}

func (u Update) GetBuildHistory() int {
	return u.BuildHistory
}

func (u Update) SetBuildHistory(buildHistory int) Task {
	u.BuildHistory = buildHistory
	return u
}

func (u Update) GetNotifications() Notifications {
	return u.Notifications
}

func (u Update) SetNotifications(notifications Notifications) Task {
	u.Notifications = notifications
	return u
}

func (u Update) SetTimeout(timeout string) Task {
	u.Timeout = timeout
	return u
}

func (u Update) SetName(name string) Task {
	return u
}

func (u Update) MarshalYAML() (interface{}, error) {
	u.Type = "update"
	return u, nil
}

func (Update) ReadsFromArtifacts() bool {
	return false
}

func (Update) GetAttempts() int {
	return 1
}

func (Update) SavesArtifacts() bool {
	return false
}

func (Update) SavesArtifactsOnFailure() bool {
	return false
}

func (Update) IsManualTrigger() bool {
	return false
}

func (Update) NotifiesOnSuccess() bool {
	return false
}

func (u Update) GetTimeout() string {
	return u.Timeout
}

func (Update) GetName() string {
	return "update"
}
