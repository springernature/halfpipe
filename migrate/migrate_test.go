package migrate

import (
	"testing"

	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

type MockController struct {
	processFunc func(man manifest.Manifest) halfpipe.Response
}

func (m MockController) DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest, err error) {
	panic("implement me")
}

func (m MockController) Process(man manifest.Manifest) halfpipe.Response {
	return m.processFunc(man)
}

func TestHappyPath(t *testing.T) {
	t.Run("not migrated", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
				manifest.TimerTrigger{},
			},
			Tasks: manifest.TaskList{
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Run{},
						manifest.Run{},
					},
				},
			},
		}

		mockController := MockController{
			processFunc: func(man manifest.Manifest) (response halfpipe.Response) {
				return
			},
		}

		parseFunc := func(manifestYaml string) (man manifest.Manifest, errs []error) {
			return
		}

		renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
			return
		}

		m := NewMigrator(mockController, parseFunc, renderFunc)

		migrated, _, lintResult, updated, err := m.Migrate(man)
		assert.False(t, lintResult.HasErrors())
		assert.False(t, updated)
		assert.NoError(t, err)
		assert.Equal(t, man, migrated)

	})
}
