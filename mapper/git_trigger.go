package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type gitTriggerMapper struct {
}

func (g gitTriggerMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original
	updated.Triggers = g.updateGitTrigger(updated.Triggers)
	return updated, nil
}

func (g gitTriggerMapper) updateGitTrigger(triggerList manifest.TriggerList) (updated manifest.TriggerList) {
	for _, trigger := range triggerList {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			for i, path := range trigger.WatchedPaths {
				if path == "." {
					trigger.WatchedPaths[i] = trigger.BasePath
				}
			}
			updated = append(updated, trigger)
		default:
			updated = append(updated, trigger)
		}
	}
	return
}
func NewGitTriggerMapper() Mapper {
	return gitTriggerMapper{}
}
