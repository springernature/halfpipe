package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testOutputDefaulter struct {
	apply func(original manifest.Manifest) (updated manifest.Manifest)
}

func (t testOutputDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	return t.apply(original)
}

type testTriggersDefaulter struct {
	apply func(original manifest.TriggerList, defaults Defaults, man manifest.Manifest) (updated manifest.TriggerList)
}

type testTasksDefaulter struct {
	apply func(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList)
}

func (t testTriggersDefaulter) Apply(original manifest.TriggerList, defaults Defaults, man manifest.Manifest) (updated manifest.TriggerList) {
	return t.apply(original, defaults, man)
}

func (t testTasksDefaulter) Apply(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList) {
	return t.apply(original, defaults, man)
}

func (t testTasksEnvVarsDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
	return t.apply(original, defaults)
}

func TestCallsOutToDefaulters(t *testing.T) {
	expectedTriggers := manifest.TriggerList{
		manifest.TimerTrigger{},
		manifest.GitTrigger{},
	}

	expectedTasks := manifest.TaskList{
		manifest.Run{},
		manifest.DockerPush{},
	}

	defaults := Defaults{
		triggersDefaulter: testTriggersDefaulter{apply: func(original manifest.TriggerList, defaults Defaults, man manifest.Manifest) (updated manifest.TriggerList) {
			return expectedTriggers
		}},
		tasksDefaulter: testTasksDefaulter{apply: func(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList) {
			return expectedTasks
		}},
		outputDefaulter: testOutputDefaulter{apply: func(original manifest.Manifest) (updated manifest.Manifest) {
			return original
		}},
	}

	assert.Equal(t, manifest.Manifest{Triggers: expectedTriggers, Tasks: expectedTasks}, defaults.Apply(manifest.Manifest{}))
}

func TestApplyFeatureToggleDefaults(t *testing.T) {
	defaults := Defaults{
		triggersDefaulter: testTriggersDefaulter{apply: func(original manifest.TriggerList, defaults Defaults, man manifest.Manifest) (updated manifest.TriggerList) {
			return manifest.TriggerList{}
		}},
		tasksDefaulter: testTasksDefaulter{apply: func(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList) {
			return manifest.TaskList{}
		}},
		outputDefaulter: testOutputDefaulter{apply: func(original manifest.Manifest) (updated manifest.Manifest) {
			return original
		}},
	}

	man := manifest.Manifest{
		FeatureToggles: []string{manifest.FeatureUpdatePipeline, manifest.FeatureDockerDecompose},
	}
	assert.Equal(t, man.FeatureToggles, defaults.Apply(man).FeatureToggles)
}
