package manifest

import "slices"

type Sequence struct {
	Type  string
	Tasks TaskList
}

func (s Sequence) GetNotifications() Notifications {
	panic("GetNotifications should never be used for a sequence task as we only care about sub tasks")
}

func (s Sequence) SetNotifications(notifications Notifications) Task {
	panic("SetNotifications should never be used for a sequence task as we only care about sub tasks")
}

func (s Sequence) SetTimeout(timeout string) Task {
	panic("SetTimeout should never be used for a sequence task as we only care about sub tasks")
}

func (s Sequence) SetName(name string) Task {
	panic("SetName should never be used for a sequence task as we only care about sub tasks")
}

func (s Sequence) ReadsFromArtifacts() bool {
	return slices.ContainsFunc(s.Tasks, func(t Task) bool { return t.ReadsFromArtifacts() })
}

func (s Sequence) GetAttempts() int {
	panic("GetAttempts should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) SavesArtifacts() bool {
	return slices.ContainsFunc(s.Tasks, func(t Task) bool { return t.SavesArtifacts() })
}

func (s Sequence) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(s.Tasks, func(t Task) bool { return t.SavesArtifactsOnFailure() })
}

func (s Sequence) IsManualTrigger() bool {
	panic("IsManualTrigger should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) NotifiesOnSuccess() bool {
	panic("NotifiesOnSuccess should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (p Sequence) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	panic("SetNotifyOnSuccess should never be used in the rendering for a sequence task as we only care about sub tasks")
}
func (s Sequence) GetTimeout() string {
	panic("GetTimeout should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) GetName() string {
	panic("GetName should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) GetGitHubEnvironment() GitHubEnvironment {
	panic("GetGitHubEnvironment should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) MarshalYAML() (interface{}, error) {
	s.Type = "sequence"
	return s, nil
}
