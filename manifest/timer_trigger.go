package manifest

// timer trigger runs the pipeline on a schedule. The cron expression must be
// valid; remember to specify times in UTC. See [crontab.guru] for help
// writing cron expressions.
//
// [crontab.guru]: https://crontab.guru/
type TimerTrigger struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Cron expression for the timer schedule. Times must be in UTC.
	Cron string `json:"cron,omitempty" yaml:"cron,omitempty" jsonschema:"required"`
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
