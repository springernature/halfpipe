package manifest

type TimerTrigger struct {
	Type string
	Cron string `yaml:"cron,omitempty"`
}

func (t TimerTrigger) GetTriggerAttempts() int {
	return 2
}

func (t TimerTrigger) MarshalYAML() (interface{}, error) {
	t.Type = "timer"
	return t, nil
}

func (TimerTrigger) GetTriggerName() string {
	return "cron"
}
