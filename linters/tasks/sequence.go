package tasks

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintSequenceTask(seqTask manifest.Sequence, cameFromAParallel bool) (errs []error, warnings []error) {
	if !cameFromAParallel {
		errs = append(errs, linterrors.NewInvalidField("type", "you are only allowed to use 'sequence' inside 'parallel'"))
		return errs, warnings
	}

	if len(seqTask.Tasks) == 0 {
		errs = append(errs, linterrors.NewInvalidField("tasks", "you are not allowed to use a empty 'sequence'"))
		return errs, warnings
	}

	for _, task := range seqTask.Tasks {
		switch task.(type) {
		case manifest.Sequence:
			errs = append(errs, linterrors.NewInvalidField("tasks", "a sequence task cannot contain sequence tasks"))
		}
	}

	return errs, warnings
}
