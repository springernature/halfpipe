package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintSequenceTask(seqTask manifest.Sequence, cameFromAParallel bool) (errs []error) {
	if !cameFromAParallel {
		errs = append(errs, NewErrInvalidField("type", "you are only allowed to use 'sequence' inside 'parallel'"))
		return errs
	}

	if len(seqTask.Tasks) == 0 {
		errs = append(errs, NewErrInvalidField("tasks", "you are not allowed to use a empty 'sequence'"))
		return errs
	}

	for _, task := range seqTask.Tasks {
		switch task.(type) {
		case manifest.Sequence:
			errs = append(errs, NewErrInvalidField("tasks", "a sequence task cannot contain sequence tasks"))
		}
	}

	return errs
}
