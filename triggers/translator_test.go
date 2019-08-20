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
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
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
		Triggers: manifest.TriggerList{
			manifest.CronTrigger{
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
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:          man.Repo.URI,
				BasePath:     man.Repo.BasePath,
				PrivateKey:   man.Repo.PrivateKey,
				WatchedPaths: man.Repo.WatchedPaths,
				IgnoredPaths: man.Repo.IgnoredPaths,
				GitCryptKey:  man.Repo.GitCryptKey,
				Branch:       man.Repo.Branch,
				Shallow:      man.Repo.Shallow,
			},
			manifest.CronTrigger{
				Trigger: man.CronTrigger,
			},
		},
	}
	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}

func TestDontDoAnythingIfBothGitTriggerAndRepoDefined(t *testing.T) {
	// We are catching this later in the linters

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
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: "somethingElse",
			},
		},
	}

	expectedManifest := man
	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}

func TestDontDoAnythingIfBothCronTriggersDefined(t *testing.T) {
	// We are catching this later in the linters

	man := manifest.Manifest{
		CronTrigger: "something",
		Triggers: manifest.TriggerList{
			manifest.CronTrigger{
				Trigger: "somethingElse",
			},
		},
	}

	expectedManifest := man
	assert.Equal(t, expectedManifest, NewTriggersTranslator().Translate(man))
}
