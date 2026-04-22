package manifest

import "slices"

// Parallel enables running tasks in parallel. All tasks start simultaneously;
// the group succeeds when all complete.
type Parallel struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Tasks to run in parallel. All tasks start simultaneously; the group succeeds when all complete.
	Tasks TaskList `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

func (p Parallel) GetBase() TaskBase {
	return TaskBase{}
}

func (p Parallel) SetNotifications(notifications Notifications) Task {
	panic("SetNotifications should never be used as we only care about sub tasks")
}

func (p Parallel) SetTimeout(timeout string) Task {
	panic("SetTimeout should never be used as we only care about sub tasks")
}

func (p Parallel) SetName(name string) Task {
	panic("SetName should never be used as we only care about sub tasks")
}

func (p Parallel) MarshalYAML() (any, error) {
	p.Type = "parallel"
	return p, nil
}

func (p Parallel) ReadsFromArtifacts() bool {
	return slices.ContainsFunc(p.Tasks, func(t Task) bool { return t.ReadsFromArtifacts() })
}

func (p Parallel) SavesArtifacts() bool {
	return slices.ContainsFunc(p.Tasks, func(t Task) bool { return t.SavesArtifacts() })
}

func (p Parallel) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(p.Tasks, func(t Task) bool { return t.SavesArtifactsOnFailure() })
}

func (p Parallel) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	panic("SetNotifyOnSuccess should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetName() string {
	panic("GetName should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetGitHubEnvironment() GitHubEnvironment {
	panic("GetGitHubEnvironment should never be used in the rendering for a parallel task as we only care about sub tasks")
}
