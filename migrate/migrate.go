package migrate

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/parallel"
	"github.com/springernature/halfpipe/triggers"
)

func Migrate(man manifest.Manifest) manifest.Manifest {
	man.Tasks = parallel.NewParallelMerger().MergeParallelTasks(man.Tasks)
	man = triggers.NewTriggersTranslator().Translate(man)
	return man
}
