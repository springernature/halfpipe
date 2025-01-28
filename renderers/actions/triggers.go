package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) triggers(man manifest.Manifest) (on On) {
	for _, t := range man.Triggers {
		switch trigger := t.(type) {
		case manifest.GitTrigger:
			on.Push = a.onPush(trigger, man.PipelineName())
		case manifest.TimerTrigger:
			on.Schedule = a.onSchedule(trigger)
		case manifest.DockerTrigger:
			on.RepositoryDispatch = a.onRepositoryDispatch(trigger.Image)
		}
	}
	return on
}

func (a *Actions) onPush(git manifest.GitTrigger, pipelineName string) (push Push) {
	if git.ManualTrigger {
		return push
	}
	push.Branches = Branches{git.Branch}

	for _, p := range git.WatchedPaths {

		var path string
		if strings.HasSuffix(p, "*") {
			path = p
		} else {
			path = fmt.Sprintf("%s**", p)
		}

		push.Paths = append(push.Paths, path)

	}

	if len(push.Paths) > 0 {
		push.Paths = append(push.Paths, fmt.Sprintf(".github/workflows/%s.yml", pipelineName))
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
