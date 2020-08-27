package manifest

type Parallel struct {
	Type  string
	Tasks TaskList `yaml:"tasks,omitempty"`
}

func (p Parallel) GetBuildHistory() int {
	panic("GetBuildHistory should never be used as we only care about sub tasks")
}

func (p Parallel) SetBuildHistory(buildHistory int) Task {
	panic("GetBuildHistory should never be used as we only care about sub tasks")
}

func (p Parallel) GetNotifications() Notifications {
	panic("GetNotifications should never be used as we only care about sub tasks")
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

func (p Parallel) MarshalYAML() (interface{}, error) {
	p.Type = "parallel"
	return p, nil
}

func (p Parallel) ReadsFromArtifacts() bool {
	for _, task := range p.Tasks {
		if task.ReadsFromArtifacts() {
			return true
		}
	}
	return false
}

func (Parallel) GetAttempts() int {
	panic("GetAttempts should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (p Parallel) SavesArtifacts() bool {
	for _, task := range p.Tasks {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}

func (p Parallel) SavesArtifactsOnFailure() bool {
	for _, task := range p.Tasks {
		if task.SavesArtifactsOnFailure() {
			return true
		}
	}
	return false
}

func (Parallel) IsManualTrigger() bool {
	panic("IsManualTrigger should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (p Parallel) NotifiesOnSuccess() bool {
	panic("NotifiesOnSuccess should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetTimeout() string {
	panic("GetTimeout should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetName() string {
	panic("GetName should never be used in the rendering for a parallel task as we only care about sub tasks")
}
