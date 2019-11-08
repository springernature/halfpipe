package manifest

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
	man, errs := Parse(``)
	assert.Empty(t, errs)
	assert.Equal(t, Manifest{}, man)
}

func TestTopLevel(t *testing.T) {
	t.Run("random weird field", func(t *testing.T) {
		yaml := `
FISHMANS_BEST_ALBUM: LONG_SEASON
`
		expected := Manifest{}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man)
	})

	t.Run("valid", func(t *testing.T) {
		yaml := `
team: TEAM
pipeline: PIPELINE
slack_channel: SLACK_CHANNEL 
`
		expected := Manifest{
			Team:         "TEAM",
			Pipeline:     "PIPELINE",
			SlackChannel: "SLACK_CHANNEL",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man)
	})
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

	man, errs := Parse(yaml)
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

	man, errs := Parse(yaml)
	assert.Empty(t, errs)
	assert.Equal(t, expected, man)
}

func TestTriggers(t *testing.T) {
	t.Run("empty trigger type", func(t *testing.T) {
		yaml := `
triggers: 
- branch: simon
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[0].type", errs)
	})

	t.Run("bad trigger type", func(t *testing.T) {
		yaml := `
triggers: 
- type: git
- type: bad
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[1].type", errs)
	})

	t.Run("bad field in trigger", func(t *testing.T) {
		yaml := `
triggers:
- type: git
  thisFieldDoesNotExist: yeah
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "triggers[0].thisFieldDoesNotExist", errs)
	})

	t.Run("bad type in trigger", func(t *testing.T) {
		yaml := `
triggers:
- type: docker
- type: git
  manual_trigger: yesPlz`

		_, errs := Parse(yaml)
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

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Triggers)
	})
}

func TestTasks(t *testing.T) {
	t.Run("empty task type", func(t *testing.T) {
		yaml := `
tasks: 
- something: simon
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "tasks[0].type", errs)
	})

	t.Run("bad task type", func(t *testing.T) {
		yaml := `
tasks:
- type: run
- type: bad
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "tasks[1].type", errs)
	})
	//
	t.Run("bad field in trigger", func(t *testing.T) {
		yaml := `
tasks:
- type: run
  thisFieldDoesNotExist: yeah
`
		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "tasks[0].thisFieldDoesNotExist", errs)
	})

	t.Run("bad type in trigger", func(t *testing.T) {
		yaml := `
tasks:
- type: run
- type: run
  manual_trigger: yesPlz`

		_, errs := Parse(yaml)
		linterrors.AssertInvalidFieldInErrors(t, "tasks[1]", errs)
	})

	t.Run("run", func(t *testing.T) {
		yaml := `
tasks:
- type: run
  name: NAME
  manual_trigger: true
  script: SCRIPT
  docker:
    image: IMAGE
    username: USERNAME
    password: PASSWORD
  privileged: true
  vars:
    VAR1: 1
    VAR2: true
    VAR3: "STR"
  save_artifacts:
  - PATH1
  - PATH2
  restore_artifacts: true
  save_artifacts_on_failure:
  - PATH3
  - PATH4
  retries: 3
  notify_on_success: true
  notifications:
    on_success:
    - c1
    - c2
    on_success_message: MSG1
    on_failure:
    - c3
    - c4
    on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := Run{
			Name:          "NAME",
			ManualTrigger: true,
			Script:        "SCRIPT",
			Docker: Docker{
				Image:    "IMAGE",
				Username: "USERNAME",
				Password: "PASSWORD",
			},
			Privileged: true,
			Vars: Vars{
				"VAR1": "1",
				"VAR2": "true",
				"VAR3": "STR",
			},
			SaveArtifacts: []string{
				"PATH1",
				"PATH2",
			},
			RestoreArtifacts: true,
			SaveArtifactsOnFailure: []string{
				"PATH3",
				"PATH4",
			},
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])

	})

	t.Run("docker-compose", func(t *testing.T) {
		yaml := `
tasks:
- type: docker-compose
  name: NAME
  command: COMMAND
  manual_trigger: true
  service: SERVICE
  compose_file: COMPOSE_FILE

  vars:
    VAR1: 1
    VAR2: true
    VAR3: STR
  save_artifacts:
  - PATH1
  - PATH2
  restore_artifacts: true
  save_artifacts_on_failure: 
  - PATH3
  - PATH4
  retries: 3
  notify_on_success: true
  notifications:
    on_success:
    - c1
    - c2
    on_success_message: MSG1
    on_failure:
    - c3
    - c4
    on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := DockerCompose{
			Name:          "NAME",
			Command:       "COMMAND",
			ManualTrigger: true,
			Service:       "SERVICE",
			ComposeFile:   "COMPOSE_FILE",

			Vars: Vars{
				"VAR1": "1",
				"VAR2": "true",
				"VAR3": "STR",
			},
			SaveArtifacts: []string{
				"PATH1",
				"PATH2",
			},
			RestoreArtifacts: true,
			SaveArtifactsOnFailure: []string{
				"PATH3",
				"PATH4",
			},
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("deploy-cf", func(t *testing.T) {
		yaml := `
tasks:
- type: deploy-cf
  name: NAME
  api: API
  space: SPACE
  org: ORG
  username: USERNAME
  password: PASSWORD
  manifest: MANIFEST
  test_domain: TEST_DOMAIN
  pre_start:
  - PS1
  deploy_artifact: DEPLOY_ARTIFACT
  pre_promote:
  - type: run
    name: PP1
  - type: run
    name: PP2
  manual_trigger: true
  vars:
   VAR1: 1
   VAR2: true
   VAR3: STR
  retries: 3
  notify_on_success: true
  notifications:
   on_success:
   - c1
   - c2
   on_success_message: MSG1
   on_failure:
   - c3
   - c4
   on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := DeployCF{
			Name:       "NAME",
			API:        "API",
			Space:      "SPACE",
			Org:        "ORG",
			Username:   "USERNAME",
			Password:   "PASSWORD",
			Manifest:   "MANIFEST",
			TestDomain: "TEST_DOMAIN",
			PreStart:   []string{"PS1"},
			PrePromote: TaskList{
				Run{Name: "PP1"},
				Run{Name: "PP2"},
			},
			DeployArtifact: "DEPLOY_ARTIFACT",
			ManualTrigger:  true,
			Vars: Vars{
				"VAR1": "1",
				"VAR2": "true",
				"VAR3": "STR",
			},
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("docker-push", func(t *testing.T) {
		yaml := `
tasks:
- type: docker-push
  name: NAME
  username: USERNAME
  password: PASSWORD
  image: IMAGE
  dockerfile_path: DOCKER_FILE_PATH
  build_path: BUILD_PATH

  manual_trigger: true
  vars:
   VAR1: 1
   VAR2: true
   VAR3: STR
  retries: 3
  notify_on_success: true
  notifications:
   on_success:
   - c1
   - c2
   on_success_message: MSG1
   on_failure:
   - c3
   - c4
   on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := DockerPush{
			Name:           "NAME",
			Username:       "USERNAME",
			Password:       "PASSWORD",
			Image:          "IMAGE",
			DockerfilePath: "DOCKER_FILE_PATH",
			BuildPath:      "BUILD_PATH",

			ManualTrigger: true,
			Vars: Vars{
				"VAR1": "1",
				"VAR2": "true",
				"VAR3": "STR",
			},
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("consumer-integration-test", func(t *testing.T) {
		yaml := `
tasks:
- type: consumer-integration-test
  name: NAME
  consumer: CONSUMER
  consumer_host: CONSUMER_HOST
  git_clone_options: GIT_CLONE_OPTIONS
  provider_host: PROVIDER_HOST
  script: SCRIPT
  docker_compose_service: DOCKER_COMPOSE_SERVICE

  vars:
   VAR1: 1
   VAR2: true
   VAR3: STR
  retries: 3
  notify_on_success: true
  notifications:
   on_success:
   - c1
   - c2
   on_success_message: MSG1
   on_failure:
   - c3
   - c4
   on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := ConsumerIntegrationTest{
			Name:                 "NAME",
			Consumer:             "CONSUMER",
			ConsumerHost:         "CONSUMER_HOST",
			GitCloneOptions:      "GIT_CLONE_OPTIONS",
			ProviderHost:         "PROVIDER_HOST",
			Script:               "SCRIPT",
			DockerComposeService: "DOCKER_COMPOSE_SERVICE",

			Vars: Vars{
				"VAR1": "1",
				"VAR2": "true",
				"VAR3": "STR",
			},
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("deploy-ml-zip", func(t *testing.T) {
		yaml := `
tasks:
- type: deploy-ml-zip
  name: NAME
  deploy_zip: DEPLOY_ZIP
  app_name: APP_NAME
  app_version: APP_VERSION
  targets:
  - T1
  - T2
  use_build_version: true

  manual_trigger: true
  retries: 3
  notify_on_success: true
  notifications:
   on_success:
   - c1
   - c2
   on_success_message: MSG1
   on_failure:
   - c3
   - c4
   on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := DeployMLZip{
			Name:            "NAME",
			DeployZip:       "DEPLOY_ZIP",
			AppName:         "APP_NAME",
			AppVersion:      "APP_VERSION",
			Targets:         []string{"T1", "T2"},
			UseBuildVersion: true,

			ManualTrigger:   true,
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("deploy-ml-modules", func(t *testing.T) {
		yaml := `
tasks:
- type: deploy-ml-modules
  name: NAME
  ml_modules_version: ML_MODULES_VERSION
  app_name: APP_NAME
  app_version: APP_VERSION
  targets:
  - T1
  - T2
  use_build_version: true

  manual_trigger: true
  retries: 3
  notify_on_success: true
  notifications:
   on_success:
   - c1
   - c2
   on_success_message: MSG1
   on_failure:
   - c3
   - c4
   on_failure_message: MSG2
  timeout: TIMEOUT
`
		expected := DeployMLModules{
			Name:             "NAME",
			MLModulesVersion: "ML_MODULES_VERSION",
			AppName:          "APP_NAME",
			AppVersion:       "APP_VERSION",
			Targets:          []string{"T1", "T2"},
			UseBuildVersion:  true,

			ManualTrigger:   true,
			Retries:         3,
			NotifyOnSuccess: true,
			Notifications: Notifications{
				OnSuccess:        []string{"c1", "c2"},
				OnSuccessMessage: "MSG1",
				OnFailure:        []string{"c3", "c4"},
				OnFailureMessage: "MSG2",
			},
			Timeout: "TIMEOUT",
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("parallel", func(t *testing.T) {
		yaml := `
tasks:
- type: parallel
  tasks:
  - type: run
    name: p1
  - type: docker-push
    name: p2
`
		expected := Parallel{
			Tasks: TaskList{
				Run{Name: "p1"},
				DockerPush{Name: "p2"},
			},
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})

	t.Run("sequence", func(t *testing.T) {
		yaml := `
tasks:
- type: sequence
  tasks:
  - type: run
    name: s1
  - type: docker-push
    name: s2
`
		expected := Sequence{
			Tasks: TaskList{
				Run{Name: "s1"},
				DockerPush{Name: "s2"},
			},
		}

		man, errs := Parse(yaml)
		assert.Empty(t, errs)
		assert.Equal(t, expected, man.Tasks[0])
	})
}
