package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestTriggersDefaulter struct {
	apply func(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList)
}

func (t TestTriggersDefaulter) Apply(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList) {
	return t.apply(original, defaults, man)
}

type TestTasksRenamer struct {
	apply func(original manifest.TaskList) (updated manifest.TaskList)
}

func (t TestTasksRenamer) Apply(original manifest.TaskList) (updated manifest.TaskList) {
	return t.apply(original)
}

type TestTasksDefaulter struct {
	apply func(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList)
}

func (t TestTasksDefaulter) Apply(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList) {
	return t.apply(original, defaults)
}

func TestUpdatePipeline(t *testing.T) {
	defaults := DefaultsNew{
		triggersDefaulter: TestTriggersDefaulter{apply: func(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList) {
			return original
		}},
		tasksRenamer: TestTasksRenamer{apply: func(original manifest.TaskList) (updated manifest.TaskList) {
			return original
		}},
		tasksDefaulter: TestTasksDefaulter{apply: func(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList) {
			return original
		}},
	}

	t.Run("doesnt do anything when feature toggle is not enabled", func(t *testing.T) {
		assert.Equal(t, manifest.Manifest{}, defaults.Apply(manifest.Manifest{}))
	})

	t.Run("adds update job as first job if feature toggle is enabled", func(t *testing.T) {
		man := manifest.Manifest{
			FeatureToggles: manifest.FeatureToggles{
				manifest.FeatureUpdatePipeline,
			},
		}

		expected := manifest.Manifest{
			FeatureToggles: manifest.FeatureToggles{
				manifest.FeatureUpdatePipeline,
			},
			Tasks: manifest.TaskList{
				manifest.Update{},
			},
		}
		assert.Equal(t, expected, defaults.Apply(man))
	})
}

func TestUpdateTriggers(t *testing.T) {
	t.Run("calls out to the triggers updated", func(t *testing.T) {
		expectedTriggers := manifest.TriggerList{
			manifest.PipelineTrigger{Status: "kehe"},
		}

		defaults := DefaultsNew{
			triggersDefaulter: TestTriggersDefaulter{apply: func(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList) {
				return expectedTriggers
			}},
			tasksRenamer: TestTasksRenamer{apply: func(original manifest.TaskList) (updated manifest.TaskList) {
				return original
			}},
			tasksDefaulter: TestTasksDefaulter{apply: func(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList) {
				return original
			}},
		}

		assert.Equal(t, expectedTriggers, defaults.Apply(manifest.Manifest{}).Triggers)
	})
}

func TestUpdateNames(t *testing.T) {
	t.Run("calls out to the triggers updated", func(t *testing.T) {
		expectedJobs := manifest.TaskList{
			manifest.Run{Name: "asdf"},
			manifest.DeployCF{Name: "kehe"},
		}

		defaults := DefaultsNew{
			triggersDefaulter: TestTriggersDefaulter{apply: func(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList) {
				return original
			}},
			tasksRenamer: TestTasksRenamer{apply: func(original manifest.TaskList) (updated manifest.TaskList) {
				return expectedJobs
			}},
			tasksDefaulter: TestTasksDefaulter{apply: func(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList) {
				return expectedJobs
			}},
		}

		assert.Equal(t, expectedJobs, defaults.Apply(manifest.Manifest{}).Tasks)
	})
}
