package linters

import (
	"errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyResolver = func() (string, error) {
	return "", nil
}

func TestActionsLinter_UnsupportedTriggers(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Triggers: manifest.TriggerList{
			manifest.DockerTrigger{},
			manifest.GitTrigger{},
			manifest.PipelineTrigger{},
			manifest.TimerTrigger{},
		},
	}

	errs := NewActionsLinter(emptyResolver).Lint(man).Issues
	assertContainsError(t, errs, ErrUnsupportedPipelineTrigger)
}

func TestActionsLinter_UnsupportedGitTriggerOptions(t *testing.T) {
	t.Run("When branch resolver returns an error", func(t *testing.T) {
		man := manifest.Manifest{
			Platform: "actions",
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{},
			},
		}

		e := errors.New("Meeehp")
		errs := NewActionsLinter(func() (string, error) {
			return "", e
		}).Lint(man).Issues

		assertContainsError(t, errs, e)
	})

	t.Run("When uri is the same as resolved uri", func(t *testing.T) {
		uri := "keheYoYo"
		man := manifest.Manifest{
			Platform: "actions",
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{
					URI:           uri,
					WatchedPaths:  []string{"watch"},
					IgnoredPaths:  []string{"ignore"},
					GitCryptKey:   "key",
					Branch:        "branch",
					Shallow:       false,
					ManualTrigger: true,
				},
			},
		}

		errs := NewActionsLinter(func() (string, error) {
			return uri, nil
		}).Lint(man).Issues
		assertNotContainsError(t, errs, ErrUnsupportedGitUri)
	})

	t.Run("WHen uri is different and private key is set", func(t *testing.T) {
		man := manifest.Manifest{
			Platform: "actions",
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{
					URI:           "uri",
					PrivateKey:    "key",
					WatchedPaths:  []string{"watch"},
					IgnoredPaths:  []string{"ignore"},
					GitCryptKey:   "key",
					Branch:        "branch",
					Shallow:       false,
					ManualTrigger: true,
				},
			},
		}

		errs := NewActionsLinter(emptyResolver).Lint(man).Issues
		assertContainsError(t, errs, ErrUnsupportedGitPrivateKey)
		assertContainsError(t, errs, ErrUnsupportedGitUri)
	})
}

func TestActionsLinter_UnsupportedTaskOptions(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Tasks: manifest.TaskList{
			manifest.DockerPush{ManualTrigger: true},
			manifest.Run{ManualTrigger: true},
			manifest.DeployCF{
				ManualTrigger: true,
				PrePromote:    manifest.TaskList{manifest.Run{}},
				Rolling:       true,
			},
		},
	}
	errs := NewActionsLinter(emptyResolver).Lint(man).Issues

	if assert.Len(t, errs, 4) {
		assert.Contains(t, errs[0].Error(), "manual_trigger")
		assert.Contains(t, errs[1].Error(), "manual_trigger")
		assert.Contains(t, errs[2].Error(), "manual_trigger")
		assert.Contains(t, errs[3].Error(), "rolling")
	}
}

func TestActionsLinter_PreventCircularTriggers(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Triggers: manifest.TriggerList{
			manifest.DockerTrigger{
				Image: "the-same-image",
			},
		},
		Tasks: manifest.TaskList{
			manifest.DockerPush{
				Image: "the-same-image",
			},
		},
	}

	errs := NewActionsLinter(emptyResolver).Lint(man).Issues
	assertContainsError(t, errs, ErrDockerTriggerLoop)
}
