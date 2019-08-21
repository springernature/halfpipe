package manifest

type TimerTrigger struct {
	Type string
	Cron string `json:"cron,omitempty" yaml:"cron,omitempty"`
}

func (t TimerTrigger) MarshalYAML() (interface{}, error) {
	t.Type = "timer"
	return t, nil
}

func (TimerTrigger) GetTriggerName() string {
	return "cron"
}
