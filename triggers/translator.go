package triggers

import (
	"github.com/springernature/halfpipe/manifest"
	"reflect"
)

type Translator struct {
}

func NewTriggersTranslator() Translator {
	return Translator{}
}

func (Translator) repoToGitTrigger(repo manifest.Repo) manifest.GitTrigger {
	return manifest.GitTrigger{
		URI:          repo.URI,
		BasePath:     repo.BasePath,
		PrivateKey:   repo.PrivateKey,
		WatchedPaths: repo.WatchedPaths,
		IgnoredPaths: repo.IgnoredPaths,
		GitCryptKey:  repo.GitCryptKey,
		Branch:       repo.Branch,
		Shallow:      repo.Shallow,
	}
}

func (Translator) cronTriggerToCronTriggerType(cronTrigger string) manifest.CronTrigger {
	return manifest.CronTrigger{
		Trigger: cronTrigger,
	}
}

func (t Translator) numTriggers(triggers manifest.TriggerList) (numGitTriggers, numCronTriggers int) {
	for _, trigger := range triggers {
		switch trigger.(type) {
		case manifest.GitTrigger:
			numGitTriggers++
		case manifest.CronTrigger:
			numCronTriggers++
		}
	}
	return
}

func (t Translator) Translate(man manifest.Manifest) manifest.Manifest {
	updatedManifest := man

	numGitTriggers, numCronTriggers := t.numTriggers(man.Triggers)

	if !reflect.DeepEqual(man.Repo, manifest.Repo{}) && numGitTriggers == 0 {
		updatedManifest.Repo = manifest.Repo{}
		updatedManifest.Triggers = append(updatedManifest.Triggers, t.repoToGitTrigger(man.Repo))
	}

	if man.CronTrigger != "" && numCronTriggers == 0 {
		updatedManifest.CronTrigger = ""
		updatedManifest.Triggers = append(updatedManifest.Triggers, t.cronTriggerToCronTriggerType(man.CronTrigger))
	}

	return updatedManifest
}
