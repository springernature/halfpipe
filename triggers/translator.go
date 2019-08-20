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

func (t Translator) Translate(man manifest.Manifest) manifest.Manifest {
	updatedManifest := man

	if !reflect.DeepEqual(man.Repo, manifest.Repo{}) {
		updatedManifest.Triggers = append(updatedManifest.Triggers, t.repoToGitTrigger(man.Repo))
		updatedManifest.Repo = manifest.Repo{}
	}

	if man.CronTrigger != "" {
		updatedManifest.Triggers = append(updatedManifest.Triggers, t.cronTriggerToCronTriggerType(man.CronTrigger))
		updatedManifest.CronTrigger = ""
	}

	return updatedManifest
}
