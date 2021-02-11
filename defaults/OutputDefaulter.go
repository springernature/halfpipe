package defaults

import "github.com/springernature/halfpipe/manifest"

type outputDefaulter struct {
}

func (o outputDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original
	if updated.Platform == "" {
		updated.Platform = "concourse"
	}

	return
}

func NewOutputDefaulter() OutputDefaulter {
	return outputDefaulter{}
}
