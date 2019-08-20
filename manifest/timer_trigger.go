package manifest

type TimerTrigger struct {
	Type string
	Cron string `json:"cron,omitempty" yaml:"cron,omitempty"`
}

func (TimerTrigger) GetTriggerName() string {
	return "cron"
}
