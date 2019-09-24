package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeqMustComeFromAParallelTask(t *testing.T) {
	errs, warnings := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, false)

	helpers.AssertInvalidFieldInErrors(t, "type", errs)
	assert.Empty(t, warnings)
}

func TestSeqIsAtLeastOne(t *testing.T) {
	t.Run("errors with empty sequence", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{}, true)

		helpers.AssertInvalidFieldInErrors(t, "tasks", errs)
		assert.Empty(t, warnings)
	})

	t.Run("warns with one task", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}}}, true)

		assert.Empty(t, errs)
		helpers.AssertInvalidFieldInErrors(t, "tasks", warnings)
	})

	t.Run("ok with two task", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, true)

		assert.Empty(t, errs)
		assert.Empty(t, warnings)
	})
}

func TestSeqDoesNotContainOtherSeqsOrParallels(t *testing.T) {
	t.Run("errors when sequence contains sequence", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{
			Type: "",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Sequence{},
			},
		}, true)
		assert.Len(t, errs, 1)
		assert.Len(t, warnings, 0)
		helpers.AssertInvalidFieldInErrors(t, "tasks", errs)
	})

	t.Run("errors when sequence contains parallel", func(t *testing.T) {
		errs, warnings := LintSequenceTask(manifest.Sequence{
			Type: "",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Parallel{},
			},
		}, true)
		assert.Len(t, errs, 1)
		assert.Len(t, warnings, 0)
		helpers.AssertInvalidFieldInErrors(t, "tasks", errs)
	})
}
