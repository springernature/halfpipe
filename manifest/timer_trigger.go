package manifest

type TimerTrigger struct {
	Type string
	Cron string `json:"cron,omitempty" yaml:"cron,omitempty"`
}

func (t TimerTrigger) GetTriggerAttempts() int {
	return 2
}

func (t TimerTrigger) MarshalYAML() (any, error) {
	t.Type = "timer"
	return t, nil
}

func (TimerTrigger) GetTriggerName() string {
	return "cron"
}
