package actions

import (
	"github.com/springernature/halfpipe/manifest"
)

type Renderer interface {
	Render(manifest manifest.Manifest) Actions
}

type pipeline struct {
}

func NewRenderer() pipeline {
	return pipeline{}
}

func (p pipeline) Render(manifest manifest.Manifest) (actions Actions) {
	actions.Name = manifest.Pipeline
	actions.On = p.triggers(manifest.Triggers)

	return
}

func (p pipeline) triggers(triggers []manifest.Trigger) (on On) {
	for _, trigger := range triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			on.Push.Branches = []string{trigger.Branch}
			on.Push.Paths = trigger.WatchedPaths
		}
	}
	return
}
