package manifest

type Update struct {
	Type     string
	TagRepo  bool
	TaskBase `yaml:",inline"`
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

func (u Update) MarshalYAML() (any, error) {
	u.Type = "update"
	return u, nil
}

func (Update) ReadsFromArtifacts() bool {
	return false
}

func (Update) SavesArtifacts() bool {
	return false
}

func (Update) SavesArtifactsOnFailure() bool {
	return false
}

func (u Update) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	return u
}

func (Update) GetName() string {
	return "update"
}
