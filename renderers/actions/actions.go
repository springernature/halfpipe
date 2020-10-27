package actions

import (
	"github.com/springernature/halfpipe/manifest"
)

type Actions struct{}

func NewActions() Actions {
	return Actions{}
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.onTriggers(man.Triggers)
	return w.asYAML()
}

func (a Actions) onTriggers(triggers manifest.TriggerList) (on On) {
	for _, trigger := range triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			on.Push = a.onPush(trigger)
		case manifest.TimerTrigger:
			on.Schedule = a.onSchedule(trigger)
		}
	}
	return on
}

func (a Actions) onPush(git manifest.GitTrigger) (push Push) {
	if git.ManualTrigger {
		return push
	}

	push.Branches = Branches{git.Branch}
	push.Paths = git.WatchedPaths

	for _, p := range git.IgnoredPaths {
		push.Paths = append(push.Paths, "!"+p)
	}
	return push
}

func (a Actions) onSchedule(timer manifest.TimerTrigger) []Cron {
	return []Cron{{timer.Cron}}
}
