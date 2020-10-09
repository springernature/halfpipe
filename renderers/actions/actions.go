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
	w.On = On{
		Push: a.OnPush(man),
	}
	w.Jobs = a.Jobs(man)

	return w.asYAML()
}

func (a Actions) OnPush(man manifest.Manifest) Push {
	gitTrigger := man.Triggers.GetGitTrigger()
	return Push{
		Branches: Branches{gitTrigger.Branch},
		Paths:    gitTrigger.WatchedPaths,
	}
}

func (a Actions) Jobs(man manifest.Manifest) []Job {
	return []Job{{
		Name: "foo",
	}}
}
