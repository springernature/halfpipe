package defaults

import "github.com/springernature/halfpipe/manifest"

type triggersDefaulter struct {
	gitTriggerDefaulter      func(original manifest.GitTrigger, defaults DefaultsNew) (updated manifest.GitTrigger)
	timerTriggerDefaulter    func(original manifest.TimerTrigger, defaults DefaultsNew) (updated manifest.TimerTrigger)
	pipelineTriggerDefaulter func(original manifest.PipelineTrigger, defaults DefaultsNew, man manifest.Manifest) (updated manifest.PipelineTrigger)
	dockerTriggerDefaulter   func(original manifest.DockerTrigger, defaults DefaultsNew) (updated manifest.DockerTrigger)
}

func NewTriggersDefaulter() TriggersDefaulter {
	return triggersDefaulter{
		timerTriggerDefaulter:    defaultTimerTrigger,
		dockerTriggerDefaulter:   defaultDockerTrigger,
		pipelineTriggerDefaulter: defaultPipelineTrigger,
		gitTriggerDefaulter:      defaultGitTrigger,
	}
}

func (t triggersDefaulter) Apply(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList) {
	for _, trigger := range original {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			updated = append(updated, t.gitTriggerDefaulter(trigger, defaults))
		case manifest.TimerTrigger:
			updated = append(updated, t.timerTriggerDefaulter(trigger, defaults))
		case manifest.PipelineTrigger:
			updated = append(updated, t.pipelineTriggerDefaulter(trigger, defaults, man))
		case manifest.DockerTrigger:
			updated = append(updated, t.dockerTriggerDefaulter(trigger, defaults))
		default:
			panic("unknown")
		}
	}
	return
}
