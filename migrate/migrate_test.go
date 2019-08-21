package migrate

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTriggers(t *testing.T) {
	t.Run("cron trigger", func(t *testing.T) {
		t.Run("when there is no cron trigger, dont do anything", func(t *testing.T) {
			man := manifest.Manifest{}

			assert.Equal(t, man, Migrate(man))
		})

		t.Run("when there is a cron trigger, convert it", func(t *testing.T) {
			cron := "* * * * *"
			man := manifest.Manifest{
				CronTrigger: cron,
			}

			expected := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.TimerTrigger{
						Cron: cron,
					},
				},
			}

			assert.Equal(t, expected, Migrate(man))
		})

		t.Run("when both cron and timer defined, dont do anything as this will be caught in linter", func(t *testing.T) {
			man := manifest.Manifest{
				CronTrigger: "* * * * *",
				Triggers: manifest.TriggerList{
					manifest.TimerTrigger{
						Cron: "wakawakaHazza",
					},
				},
			}

			assert.Equal(t, man, Migrate(man))
		})
	})
	t.Run("git trigger", func(t *testing.T) {
		t.Run("when there is no git trigger, dont do anything", func(t *testing.T) {
			man := manifest.Manifest{}

			assert.Equal(t, man, Migrate(man))
		})

		t.Run("when there is no cron trigger, dont do anything", func(t *testing.T) {
			uri := "blurgh"
			man := manifest.Manifest{
				Repo: manifest.Repo{
					URI: uri,
				},
			}

			expected := manifest.Manifest{
				Triggers: manifest.TriggerList{
					manifest.GitTrigger{
						URI: uri,
					},
				},
			}

			assert.Equal(t, expected, Migrate(man))
		})

		t.Run("when both cron and timer defined, dont do anything as this will be caught in linter", func(t *testing.T) {
			man := manifest.Manifest{
				Repo: manifest.Repo{
					URI: "blurgh",
				},
				Triggers: manifest.TriggerList{
					manifest.GitTrigger{
						URI: "wakawakaHazza",
					},
				},
			}

			assert.Equal(t, man, Migrate(man))
		})

	})
}

func TestParallelMerger(t *testing.T) {
	t.Run("no groups", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{},
			},
		}

		assert.Equal(t, man, Migrate(man))
	})

	t.Run("groups", func(t *testing.T) {
		// No need to test to much as this is tested properly in the parallel.Merger
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{Parallel: "asdf"},
				manifest.Run{Parallel: "asdf"},
				manifest.Run{},
			},
		}

		expected := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Parallel{
					Tasks:manifest.TaskList{
						manifest.Run{},
						manifest.Run{},
					},
				},
				manifest.Run{},
			},
		}

		assert.Equal(t, expected, Migrate(man))

	})
}
