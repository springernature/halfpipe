package manifest

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
	for _, task := range s.Tasks {
		if task.ReadsFromArtifacts() {
			return true
		}
	}
	return false
}

func (s Sequence) GetAttempts() int {
	panic("GetAttempts should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) SavesArtifacts() bool {
	for _, task := range s.Tasks {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}

func (s Sequence) SavesArtifactsOnFailure() bool {
	for _, task := range s.Tasks {
		if task.SavesArtifactsOnFailure() {
			return true
		}
	}
	return false
}

func (s Sequence) IsManualTrigger() bool {
	panic("IsManualTrigger should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) NotifiesOnSuccess() bool {
	for _, task := range s.Tasks {
		if task.NotifiesOnSuccess() {
			return true
		}
	}
	return false
}

func (s Sequence) GetTimeout() string {
	panic("GetTimeout should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) GetName() string {
	panic("GetName should never be used in the rendering for a sequence task as we only care about sub tasks")
}

func (s Sequence) MarshalYAML() (interface{}, error) {
	s.Type = "sequence"
	return s, nil
}
