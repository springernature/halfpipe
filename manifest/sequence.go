package manifest

import "slices"

// Sequence enables running tasks in sequence within a parallel group. It can
// only be used inside a parallel task.
type Sequence struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Tasks to run in sequence within a parallel group. Can only be used inside a parallel task.
	Tasks TaskList `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

func (p Sequence) GetBase() TaskBase {
	return TaskBase{}
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

func (s Sequence) SavesArtifacts() bool {
	return slices.ContainsFunc(s.Tasks, func(t Task) bool { return t.SavesArtifacts() })
}

func (s Sequence) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(s.Tasks, func(t Task) bool { return t.SavesArtifactsOnFailure() })
}

func (p Sequence) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	panic("SetNotifyOnSuccess should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) GetName() string {
	panic("GetName should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) MarshalYAML() (any, error) {
	s.Type = "sequence"
	return s, nil
}
