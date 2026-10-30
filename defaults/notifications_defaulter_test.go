package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestNotificationsDefaulterSetsDefaultFailureIfEmpty(t *testing.T) {
	expected := manifest.Manifest{
		Team: "myteam",
		Notifications: manifest.Notifications{
			Failure: manifest.NotificationChannels{{Teams: config.PlatformAPIMessageURL + "?team=myteam"}},
		},
	}
	assert.Equal(t, expected, NewNotificationsDefaulter().Apply(manifest.Manifest{Team: "myteam"}))
}

func TestNotificationsDefaulterDoesNothingIfFailureSet(t *testing.T) {
	original := manifest.Manifest{
		Team: "myteam",
		Notifications: manifest.Notifications{
			Failure: manifest.NotificationChannels{{Teams: "https://example.com"}},
		},
	}
	assert.Equal(t, original, NewNotificationsDefaulter().Apply(original))
}

func TestNotificationsDefaulterDoesNothingIfDisabled(t *testing.T) {
	original := manifest.Manifest{
		Team:           "myteam",
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureDisableAutomaticNotifications},
	}
	assert.Equal(t, original, NewNotificationsDefaulter().Apply(original))
}
