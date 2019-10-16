package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCallsOutCorrectly(t *testing.T) {
	expectedGitTrigger := manifest.GitTrigger{URI: "blah"}
	expectedTimerTrigger := manifest.TimerTrigger{Cron: "fdas"}
	expectedPipelineTrigger := manifest.PipelineTrigger{Pipeline: "asdf"}
	expectedDockerTrigger := manifest.DockerTrigger{Image: "asdf"}

	defaulter := triggersDefaulter{
		gitTriggerDefaulter: func(original manifest.GitTrigger, defaults Defaults) (updated manifest.GitTrigger) {
			return expectedGitTrigger
		},
		timerTriggerDefaulter: func(original manifest.TimerTrigger, defaults Defaults) (updated manifest.TimerTrigger) {
			return expectedTimerTrigger
		},
		pipelineTriggerDefaulter: func(original manifest.PipelineTrigger, defaults Defaults, man manifest.Manifest) (updated manifest.PipelineTrigger) {
			return expectedPipelineTrigger
		},
		dockerTriggerDefaulter: func(original manifest.DockerTrigger, defaults Defaults) (updated manifest.DockerTrigger) {
			return expectedDockerTrigger
		},
	}

	updated := defaulter.Apply(manifest.TriggerList{
		manifest.GitTrigger{},
		manifest.TimerTrigger{},
		manifest.PipelineTrigger{},
		manifest.DockerTrigger{},
		manifest.GitTrigger{},
		manifest.TimerTrigger{},
		manifest.PipelineTrigger{},
		manifest.DockerTrigger{},
	}, DefaultValues, manifest.Manifest{})

	expected := manifest.TriggerList{
		expectedGitTrigger,
		expectedTimerTrigger,
		expectedPipelineTrigger,
		expectedDockerTrigger,
		expectedGitTrigger,
		expectedTimerTrigger,
		expectedPipelineTrigger,
		expectedDockerTrigger,
	}

	assert.Equal(t, expected, updated)
}

func TestAddsDefaultGitTriggerIfThereIsntOneInTheTriggerList(t *testing.T) {
	expectedGitTrigger := manifest.GitTrigger{
		URI: "meehp",
	}

	timerTrigger := manifest.TimerTrigger{
		Cron: "asdf",
	}

	defaulter := triggersDefaulter{
		gitTriggerDefaulter: func(original manifest.GitTrigger, defaults Defaults) (updated manifest.GitTrigger) {
			return expectedGitTrigger
		},
		timerTriggerDefaulter: func(original manifest.TimerTrigger, defaults Defaults) (updated manifest.TimerTrigger) {
			return original
		},
	}

	input := manifest.TriggerList{
		timerTrigger,
	}

	expected := manifest.TriggerList{
		timerTrigger,
		expectedGitTrigger,
	}

	assert.Equal(t, expected, defaulter.Apply(input, DefaultValues, manifest.Manifest{}))
}
