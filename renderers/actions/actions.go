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
	return w.asYAML()
}
