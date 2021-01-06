package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func (t testTasksArtifactoryVarsDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
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
	}

	man := manifest.Manifest{
		FeatureToggles: []string{manifest.FeatureUpdatePipeline, manifest.FeatureDockerDecompose},
	}

	assert.Equal(t, man.FeatureToggles, defaults.Apply(man).FeatureToggles)
	defaults.RemoveUpdatePipelineFeature = true
	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureDockerDecompose}, defaults.Apply(man).FeatureToggles)
}
