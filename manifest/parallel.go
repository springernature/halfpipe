package manifest

type Parallel struct {
	Type  string
	Tasks TaskList `yaml:"tasks,omitempty"`
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
	panic("this should never be used in the rendering for a parallel task as we only care about sub tasks")
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
	panic("this should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (p Parallel) NotifiesOnSuccess() bool {
	for _, task := range p.Tasks {
		if task.NotifiesOnSuccess() {
			return true
		}
	}
	return false
}

func (Parallel) GetTimeout() string {
	panic("this should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetParallelGroup() ParallelGroup {
	panic("this should never be used in the rendering for a parallel task as we only care about sub tasks")
}

func (Parallel) GetName() string {
	return "No used"
}
