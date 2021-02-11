package defaults

import "github.com/springernature/halfpipe/manifest"

type outputDefaulter struct {
}

func (o outputDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original
	if updated.Output == "" {
		updated.Output = "concourse"
	}

	return
}

func NewOutputDefaulter() OutputDefaulter {
	return outputDefaulter{}
}
