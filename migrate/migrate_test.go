package migrate

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/pkg/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockController struct {
	processFunc func(man manifest.Manifest) (config atc.Config, results result.LintResults)
}

func (m MockController) Process(man manifest.Manifest) (config atc.Config, results result.LintResults) {
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
			processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
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

	t.Run("migrated", func(t *testing.T) {
		uri := "blah"
		cronTrigger := "blug"
		name1 := "asd"
		name2 := "420"
		man := manifest.Manifest{
			CronTrigger: cronTrigger,
			Repo: manifest.Repo{
				URI: uri,
			},
			Tasks: manifest.TaskList{
				manifest.Run{Name: name1, Parallel: "true"},
				manifest.Run{Name: name2, Parallel: "true"},
			},
		}

		expectedMan := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{URI: uri},
				manifest.TimerTrigger{Cron: cronTrigger},
			},
			Tasks: manifest.TaskList{
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Run{Name: name1},
						manifest.Run{Name: name2},
					},
				},
			},
		}

		mockController := MockController{
			processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
				return
			},
		}

		parseFunc := func(manifestYaml string) (man manifest.Manifest, errs []error) {
			man = expectedMan
			return
		}

		renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
			return
		}

		m := NewMigrator(mockController, parseFunc, renderFunc)

		migrated, _, lintResult, updated, err := m.Migrate(man)
		assert.False(t, lintResult.HasErrors())
		assert.True(t, updated)
		assert.NoError(t, err)
		assert.Equal(t, expectedMan, migrated)

	})

}

func TestWhenFailingToLintOriginalManifest(t *testing.T) {

	lintResults := result.LintResults{
		result.LintResult{
			Errors: []error{errors.New("Some error")},
		},
	}

	mockController := MockController{
		processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
			results = lintResults
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

	man := manifest.Manifest{}
	migrated, _, lintResult, updated, err := m.Migrate(man)

	assert.False(t, updated)
	assert.Equal(t, man, migrated)
	assert.Equal(t, lintResults, lintResult)
	assert.Equal(t, ErrLintingOriginalManifest, err)
}

func TestWhenFailingLintTheMigratedManifest(t *testing.T) {

	lintResults := result.LintResults{
		result.LintResult{
			Errors: []error{errors.New("Some error")},
		},
	}

	var callCount int
	mockController := MockController{
		processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
			callCount++

			// Second call will be with the updated manifest
			if callCount == 2 {
				results = lintResults
			}
			return
		},
	}

	parseFunc := func(manifestYaml string) (i manifest.Manifest, i2 []error) {
		return
	}

	renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
		return
	}

	m := NewMigrator(mockController, parseFunc, renderFunc)

	man := manifest.Manifest{
		CronTrigger: "some_trigger",
	}

	migrated, _, lintResult, updated, err := m.Migrate(man)

	assert.False(t, updated)
	assert.Equal(t, man, migrated)
	assert.Equal(t, lintResults, lintResult)
	assert.Equal(t, ErrLintingMigratedManifest, err)
}

func TestWhenFailingToRenderMigratedManifestToYaml(t *testing.T) {
	mockController := MockController{
		processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
			return
		},
	}

	parseFunc := func(manifestYaml string) (i manifest.Manifest, i2 []error) {
		return
	}

	expectedErr := errors.New("blurgh")
	renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
		err = expectedErr
		return
	}

	m := NewMigrator(mockController, parseFunc, renderFunc)

	man := manifest.Manifest{
		CronTrigger: "some_trigger",
	}

	updatedManifest := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "some_trigger",
			},
		},
	}

	migrated, _, lintResult, updated, err := m.Migrate(man)
	assert.False(t, updated)
	assert.Equal(t, manifest.Manifest{}, migrated)
	assert.Equal(t, result.LintResults(nil), lintResult)
	assert.Equal(t, FailedToRenderMigratedManifestToYamlErr(expectedErr, updatedManifest), err)
}

func TestWhenFailingToRParseMigratedManifestYaml(t *testing.T) {
	mockController := MockController{
		processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
			return
		},
	}

	expectedErr := []error{errors.New("blurgh")}
	parseFunc := func(manifestYaml string) (i manifest.Manifest, errs []error) {
		errs = expectedErr
		return
	}

	expectedManifestYaml := `triggers:
- type: timer
  cron: some_trigger
`
	renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
		y = []byte(expectedManifestYaml)
		return
	}

	m := NewMigrator(mockController, parseFunc, renderFunc)

	man := manifest.Manifest{
		CronTrigger: "some_trigger",
	}

	migrated, _, lintResult, updated, err := m.Migrate(man)
	assert.False(t, updated)
	assert.Equal(t, manifest.Manifest{}, migrated)
	assert.Equal(t, result.LintResults(nil), lintResult)
	assert.Equal(t, FailedToParseMigratedManifestYamlErr(expectedErr, expectedManifestYaml), err)
}

func TestWhenMigratedManifestIsNotTheSameAsTheParsedMigratedYaml(t *testing.T) {
	mockController := MockController{
		processFunc: func(man manifest.Manifest) (config atc.Config, results result.LintResults) {
			return
		},
	}

	someRandomManifest := manifest.Manifest{
		Pipeline: "wryyyyy",
	}
	parseFunc := func(manifestYaml string) (i manifest.Manifest, errs []error) {
		i = someRandomManifest
		return
	}

	renderFunc := func(manifest manifest.Manifest) (y []byte, err error) {
		return
	}

	m := NewMigrator(mockController, parseFunc, renderFunc)

	man := manifest.Manifest{
		CronTrigger: "some_trigger",
	}

	updatedManifest := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.TimerTrigger{
				Cron: "some_trigger",
			},
		},
	}

	migrated, _, lintResult, updated, err := m.Migrate(man)
	fmt.Println(err)
	assert.False(t, updated)
	assert.Equal(t, manifest.Manifest{}, migrated)
	assert.Equal(t, result.LintResults(nil), lintResult)
	assert.Equal(t, ParsedMigratedManifestAndMigratedManifestIsNotTheSameErr(updatedManifest, someRandomManifest), err)

}
