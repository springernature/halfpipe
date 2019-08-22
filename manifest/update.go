package manifest

type Update struct {
	Timeout string
}

func (u Update) MarshalYAML() (interface{}, error) {
	panic("This should never be exposed")
}

func (Update) ReadsFromArtifacts() bool {
	return false
}

func (Update) GetAttempts() int {
	return 1
}

func (Update) SavesArtifacts() bool {
	return false
}

func (Update) SavesArtifactsOnFailure() bool {
	return false
}

func (Update) IsManualTrigger() bool {
	return false
}

func (Update) NotifiesOnSuccess() bool {
	return false
}

func (u Update) GetTimeout() string {
	return u.Timeout
}

func (Update) GetParallelGroup() ParallelGroup {
	return ParallelGroup("")
}

func (Update) GetName() string {
	return "update"
}
