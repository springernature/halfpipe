package manifest

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
	man, errs := ParseV2(``)
	assert.Empty(t, errs)
	assert.Equal(t, Manifest{}, man)
}

func TestTopLevel(t *testing.T) {

	yaml := `pipeline: simon
team: asdf
triggers: 
- type: git
  branch: simon
- type: asdf
`

	ParseV2(yaml)
}

func TestArtifactConfig(t *testing.T) {
	yaml := `
artifact_config:
  bucket: BUCKET
  json_key: JSON_KEY
`
	expected := Manifest{
		ArtifactConfig: ArtifactConfig{
			Bucket:  "BUCKET",
			JSONKey: "JSON_KEY",
		},
	}

	man, errs := ParseV2(yaml)
	assert.Empty(t, errs)
	assert.Equal(t, expected, man)
}

func TestFeatureToggles(t *testing.T) {
	yaml := `
feature_toggles:
- TOGGLE1
- TOGGLE2
`
	expected := Manifest{
		FeatureToggles: []string{
			"TOGGLE1",
			"TOGGLE2",
		},
	}

	man, errs := ParseV2(yaml)
	assert.Empty(t, errs)
	assert.Equal(t, expected, man)
}

func TestTriggers(t *testing.T) {
	t.Run("empty trigger type", func(t *testing.T) {
		yaml := `
triggers: 
- branch: simon
`
		_, errs := ParseV2(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[0].type", errs)
	})

	t.Run("bad trigger type", func(t *testing.T) {
		yaml := `
triggers: 
- type: git
- type: bad
`
		_, errs := ParseV2(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[1].type", errs)
	})

	t.Run("bad field in trigger", func(t *testing.T) {
		yaml := `
triggers:
- type: git
  thisFieldDoesNotExist: yeah
`
		_, errs := ParseV2(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[0].thisFieldDoesNotExist", errs)
	})

	t.Run("bad type in trigger", func(t *testing.T) {
		yaml := `
triggers:
- type: docker
- type: git
  manual_trigger: yesPlz`

		_, errs := ParseV2(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[1]", errs)
	})

	t.Run("all triggers", func(t *testing.T) {
		yaml := `
triggers: 
- type: git
  uri: URI
  private_key: PRIVATE_KEY 
  watched_paths: 
  - WATCHED_PATH1
  - WATCHED_PATH2
  ignored_paths:
  - IGNORED_PATH1
  - IGNORED_PATH2
  git_crypt_key: GIT_CRYPT_KEY
  branch: BRANCH
  shallow: true
  manual_trigger: true
- type: docker
  image: IMAGE
  username: USERNAME
  password: PASSWORD
- type: timer
  cron: CRON_EXPR
- type: pipeline
  concourse_url: CONCOURSE_URL
  username: USERNAME
  password: PASSWORD
  team: TEAM
  pipeline: PIPELINE
  job: JOB
  status: STATUS
`
		expected := TriggerList{
			GitTrigger{
				URI:           "URI",
				PrivateKey:    "PRIVATE_KEY",
				WatchedPaths:  []string{"WATCHED_PATH1", "WATCHED_PATH2"},
				IgnoredPaths:  []string{"IGNORED_PATH1", "IGNORED_PATH2"},
				GitCryptKey:   "GIT_CRYPT_KEY",
				Branch:        "BRANCH",
				Shallow:       true,
				ManualTrigger: true,
			},
			DockerTrigger{
				Image:    "IMAGE",
				Username: "USERNAME",
				Password: "PASSWORD",
			},
			TimerTrigger{
				Cron: "CRON_EXPR",
			},
			PipelineTrigger{
				ConcourseURL: "CONCOURSE_URL",
				Username:     "USERNAME",
				Password:     "PASSWORD",
				Team:         "TEAM",
				Pipeline:     "PIPELINE",
				Job:          "JOB",
				Status:       "STATUS",
			},
		}

		man, errs := ParseV2(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Triggers)
	})
}
