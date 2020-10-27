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
	w.On = a.On(man.Triggers)

	return w.asYAML()
}

func (a Actions) On(triggers manifest.TriggerList) (on On) {
	if !triggers.HasGitTrigger() {
		return on
	}

	git := triggers.GetGitTrigger()
	on.Push = Push{
		Branches: Branches{git.Branch},
		Paths:    git.WatchedPaths,
	}

	return on
}
