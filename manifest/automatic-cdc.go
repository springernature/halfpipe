package manifest

type AutomaticCDC struct {
	Type string
	Name string `yaml:"name,omitempty"`
}

func (r AutomaticCDC) GetSecrets() map[string]string {
	return findSecrets(map[string]string{})
}

func (r AutomaticCDC) GetBuildHistory() int {
	return 0
}

func (r AutomaticCDC) SetBuildHistory(buildHistory int) Task {
	return r
}

func (r AutomaticCDC) GetNotifications() Notifications {
	return Notifications{}
}

func (r AutomaticCDC) SetNotifications(notifications Notifications) Task {
	return r
}

func (r AutomaticCDC) SetName(name string) Task {
	r.Name = name
	return r
}

func (r AutomaticCDC) SetTimeout(timeout string) Task {
	return r
}

func (r AutomaticCDC) MarshalYAML() (interface{}, error) {
	r.Type = "automatic-cdc"
	return r, nil
}

func (r AutomaticCDC) GetName() string {
	return r.Name
}

func (r AutomaticCDC) GetTimeout() string {
	return "1h"
}

func (r AutomaticCDC) NotifiesOnSuccess() bool {
	return false
}

func (r AutomaticCDC) SavesArtifactsOnFailure() bool {
	return false
}

func (r AutomaticCDC) IsManualTrigger() bool {
	return false
}

func (r AutomaticCDC) SavesArtifacts() bool {
	return false
}

func (r AutomaticCDC) ReadsFromArtifacts() bool {
	return false
}

func (r AutomaticCDC) GetAttempts() int {
	return 1
}
