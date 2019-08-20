package manifest

type CronTrigger struct {
	Type    string
	Trigger string `json:"trigger,omitempty" yaml:"trigger,omitempty"`
}

func (CronTrigger) GetTriggerName() string {
	return "cron"
}
