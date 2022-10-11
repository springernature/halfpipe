package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestSeqMustComeFromAParallelTask(t *testing.T) {
	errs := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, false)

	assertContainsError(t, errs, ErrInvalidField.WithValue("type"))
}

func TestSeqIsAtLeastOne(t *testing.T) {
	t.Run("errors with empty sequence", func(t *testing.T) {
		errs := LintSequenceTask(manifest.Sequence{}, true)

		assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
	})

	t.Run("ok with two task", func(t *testing.T) {
		errs := LintSequenceTask(manifest.Sequence{Tasks: manifest.TaskList{manifest.Run{}, manifest.Run{}}}, true)

		assert.Empty(t, errs)
	})
}

func TestSeqDoesNotContainOtherSeqsOrParallels(t *testing.T) {
	t.Run("errors when sequence contains sequence", func(t *testing.T) {
		errs := LintSequenceTask(manifest.Sequence{
			Type: "",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Sequence{},
			},
		}, true)
		assert.Len(t, errs, 1)
		assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
	})
}
