package actions

import (
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) triggers(triggers manifest.TriggerList) (on On) {
	for _, t := range triggers {
		switch trigger := t.(type) {
		case manifest.GitTrigger:
			on.Push = a.onPush(trigger)
		case manifest.TimerTrigger:
			on.Schedule = a.onSchedule(trigger)
		case manifest.DockerTrigger:
			on.RepositoryDispatch = a.onRepositoryDispatch(trigger.Image)
		}
	}
	return on
}

func (a *Actions) onPush(git manifest.GitTrigger) (push Push) {
	if git.ManualTrigger {
		return push
	}

	push.Branches = Branches{git.Branch}

	for _, p := range git.WatchedPaths {
		push.Paths = append(push.Paths, p+"**")
	}

	// if there are only ignored paths you first have to include all
	if len(git.WatchedPaths) == 0 && len(git.IgnoredPaths) > 0 {
		push.Paths = Paths{"**"}
	}

	for _, p := range git.IgnoredPaths {
		push.Paths = append(push.Paths, "!"+p+"**")
	}

	return push
}

func (a *Actions) onSchedule(timer manifest.TimerTrigger) []Cron {
	return []Cron{{timer.Cron}}
}

func (a *Actions) onRepositoryDispatch(name string) RepositoryDispatch {
	return RepositoryDispatch{
		Types: []string{"docker-push:" + name},
	}
}
