package manifest

type Seq struct {
	Type  string
	Tasks TaskList `yaml:"tasks,omitempty"`
}

func (s Seq) MarshalYAML() (interface{}, error) {
	s.Type = "seq"
	return s, nil
}

func (s Seq) ReadsFromArtifacts() bool {
	for _, task := range s.Tasks {
		if task.ReadsFromArtifacts() {
			return true
		}
	}
	return false
}

func (Seq) GetAttempts() int {
	panic("GetAttempts should never be used in the rendering for a sequential task as we only care about sub tasks")
}

func (s Seq) SavesArtifacts() bool {
	for _, task := range s.Tasks {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}

func (s Seq) SavesArtifactsOnFailure() bool {
	for _, task := range s.Tasks {
		if task.SavesArtifactsOnFailure() {
			return true
		}
	}
	return false
}

func (Seq) IsManualTrigger() bool {
	panic("IsManualTrigger should never be used in the rendering for a sequential task as we only care about sub tasks")
}

func (s Seq) NotifiesOnSuccess() bool {
	for _, task := range s.Tasks {
		if task.NotifiesOnSuccess() {
			return true
		}
	}
	return false
}

func (Seq) GetTimeout() string {
	panic("GetTimeout should never be used in the rendering for a sequential task as we only care about sub tasks")
}

func (Seq) GetParallelGroup() ParallelGroup {
	return ""
}

func (Seq) GetName() string {
	panic("GetName should never be used in the rendering for a sequential task as we only care about sub tasks")
}
