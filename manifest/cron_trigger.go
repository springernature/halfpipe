package manifest

type Cron struct {
	Type    string
	Trigger string `json:"trigger,omitempty" yaml:"trigger,omitempty"`
}

func (Cron) GetTriggerName() string {
	return "cron"
}
