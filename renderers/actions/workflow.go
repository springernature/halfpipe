package actions

import "github.com/springernature/halfpipe/manifest"

type workflow struct{}

func NewWorkflow() workflow {
	return workflow{}
}

func (w workflow) Render(man manifest.Manifest) (string, error) {
	return "workflow yaml", nil
}
