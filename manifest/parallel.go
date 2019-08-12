package manifest

type Parallel struct {
	Type  string
	Tasks TaskList
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
	panic("implement me")
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
	panic("implement me")
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
	panic("implement me")
}

func (Parallel) GetParallelGroup() ParallelGroup {
	panic("implement me")
}

func (Parallel) GetName() string {
	return "No used"
}
