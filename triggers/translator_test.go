package triggers

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTranslatesRepo(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:          "a",
			BasePath:     "b",
			PrivateKey:   "c",
			WatchedPaths: []string{"a", "b", "c"},
			IgnoredPaths: []string{"d", "e", "f"},
			GitCryptKey:  "g",
			Branch:       "h",
			Shallow:      true,
		},
	}

	expectedManifest := manifest.Manifest{
		Repo: manifest.Repo{},
		Triggers: manifest.TriggerList{
			manifest.Git{
				URI:          man.Repo.URI,
				BasePath:     man.Repo.BasePath,
				PrivateKey:   man.Repo.PrivateKey,
				WatchedPaths: man.Repo.WatchedPaths,
				IgnoredPaths: man.Repo.IgnoredPaths,
				GitCryptKey:  man.Repo.GitCryptKey,
				Branch:       man.Repo.Branch,
				Shallow:      man.Repo.Shallow,
			},
		},
	}
	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}

func TestTranslatesCron(t *testing.T) {
	man := manifest.Manifest{
		CronTrigger: "some valid cron",
	}

	expectedManifest := manifest.Manifest{
		CronTrigger: "",
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: man.CronTrigger,
			},
		},
	}

	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}

func TestTranslatesBothGitAndCron(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:          "a",
			BasePath:     "b",
			PrivateKey:   "c",
			WatchedPaths: []string{"a", "b", "c"},
			IgnoredPaths: []string{"d", "e", "f"},
			GitCryptKey:  "g",
			Branch:       "h",
			Shallow:      true,
		},
		CronTrigger: "some valid cron",
	}

	expectedManifest := manifest.Manifest{
		Repo: manifest.Repo{},
		Triggers: manifest.TriggerList{
			manifest.Git{
				URI:          man.Repo.URI,
				BasePath:     man.Repo.BasePath,
				PrivateKey:   man.Repo.PrivateKey,
				WatchedPaths: man.Repo.WatchedPaths,
				IgnoredPaths: man.Repo.IgnoredPaths,
				GitCryptKey:  man.Repo.GitCryptKey,
				Branch:       man.Repo.Branch,
				Shallow:      man.Repo.Shallow,
			},
			manifest.Cron{
				Trigger: man.CronTrigger,
			},
		},
	}
	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}
