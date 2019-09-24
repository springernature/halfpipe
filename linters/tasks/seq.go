package tasks

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintSeqTask(seqTask manifest.Seq, cameFromAParallel bool) (errs []error, warnings []error) {
	if !cameFromAParallel {
		errs = append(errs, errors.NewInvalidField("type", "You are only allowed to use a 'seq' inside a 'parallel'"))
		return
	}

	if len(seqTask.Tasks) == 0 {
		errs = append(errs, errors.NewInvalidField("tasks", "You are not allowed to use a empty 'seq'"))
		return
	}

	if len(seqTask.Tasks) == 1 {
		warnings = append(warnings, errors.NewInvalidField("tasks", "It seems unnecessary to have a single task in a seq"))
		return
	}

	for _, task := range seqTask.Tasks {
		switch task.(type) {
		case manifest.Seq:
			errs = append(errs, errors.NewInvalidField("tasks", "A seq task cannot contain seq tasks"))
		case manifest.Parallel:
			errs = append(errs, errors.NewInvalidField("tasks", "A seq task cannot contain parallel tasks"))
		}
	}

	return
}
