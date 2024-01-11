package manifest

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestEmptyManifest(t *testing.T) {
	man, errs := Parse(``)
	assert.Nil(t, errs)
	assert.Equal(t, Manifest{}, man)
}

func TestValidYaml_Everything(t *testing.T) {

	man, errs := Parse(`
team: my team
pipeline: my pipeline
slack_channel: "#ee-activity"
slack_success_message: "success"
slack_failure_message: "failure"
artifact_config:
  bucket: myBucket
  json_key: ((some.jsonKey))
triggers:
- type: git
  uri: git@github.com:..
  private_key: private-key
  watched_paths:
  - watched/dir1
  - watched/dir2
  ignored_paths:
  - ignored/dir1/**
  - README.md
  git_crypt_key: git-crypt-key
  manual_trigger: true
- type: timer
  cron: "*/10 * * * *"
- type: docker
  image: ubuntu
  username: userName
  password: password
- type: pipeline
  pipeline: a
  job: b
  status: c
feature_toggles:
- feature1
- feature2
- featureXYZ
tasks:
- type: run
  name: run task
  script: script.sh --param
  docker:
    image: golang:latest
    username: user
    password: pass
  privileged: true
  vars:
    FOO: fOo
    BAR: "1"
  save_artifacts:
  - target/dist/artifact.zip
  - README.md
  save_artifacts_on_failure:
  - test_reports
- type: docker-compose
  name: docker compose task
  compose_file: ../compose-file.yml
  vars:
    FOO: fOo
    BAR: "1"
    BAZ: true
    WRYY: 2
    THIS_IS_STRANGE:
    - a
    - b
  save_artifacts:
  - target/dist/artifact.zip
  - README.md
  save_artifacts_on_failure:
  - test_reports
- type: docker-push
  name: docker push task
  username: user
  password: pass
  image: golang:latest
  ignore_vulnerabilities: true
  notifications:
    on_success:
    - asdf
    - kehe
    on_failure:
    - kfds
    - oasdf
  vars:
    FOO: fOo
    BAR: "1"
  timeout: 1h
  tag: version
- type: deploy-cf
  name: deploy cf task
  api: cf.api
  space: cf.space
  org: cf.org
  rolling: true
  username: cf.user
  password: cf.pass
  manifest: manifest.yml
  test_domain: asdf.com
  sso_route: some.sso.route
  vars:
    FOO: fOo
    BAR: "1"
  deploy_artifact: target/dist/artifact.zip
  pre_start:
  - cf apps
  - cf events
  pre_promote:
  - type: run
    script: smoke-test.sh
    docker:
      image: golang
  - type: consumer-integration-test
    name: cdc-name
    consumer: cdc-consumer
    consumer_host: cdc-host
    script: cdc-script
    use_covenant: false
- type: deploy-katee
  name: deploy katee task
  vela_manifest: blah
  manual_trigger: false
  namespace: some-team
  timeout: 30s
  tag: latest
  notifications:
    on_success:
    - asdf
    - kehe
    on_failure:
    - kfds
    - oasdf
  vars:
    FOO: fOo
    BAR: "1"
- type: docker-compose
  name: docker compose task 2
  service: asdf
- type: consumer-integration-test
  name: cdc-name
  consumer: cdc-consumer
  consumer_host: cdc-host
  script: cdc-script
  git_clone_options: --depth 100
- type: deploy-ml-zip
  name: deploy ml zip
  app_name: app-name
  app_version: app-version
  deploy_zip: deploy-zip
  use_build_version: true
  targets:
  - target1
  - target2
  username: un
  password: pw
- type: deploy-ml-modules
  app_name: app-name
  app_version: app-version
  ml_modules_version: ml-modules-version
  password: p
  targets:
  - target1
  - target2
  build_history: 10
- type: parallel
  tasks:
  - type: run
    name: pr1
  - type: run
    name: pr2
- type: parallel
  tasks:
  - type: sequence
    tasks:
    - type: run
      name: pr1
    - type: run
      name: pr2
`)

	fmt.Print(errs)
	expected := Manifest{
		Team:     "my team",
		Pipeline: "my pipeline",
		ArtifactConfig: ArtifactConfig{
			Bucket:  "myBucket",
			JSONKey: "((some.jsonKey))",
		},
		SlackChannel:        "#ee-activity",
		SlackSuccessMessage: "success",
		SlackFailureMessage: "failure",
		FeatureToggles: FeatureToggles{
			"feature1",
			"feature2",
			"featureXYZ",
		},
		Triggers: TriggerList{
			GitTrigger{
				URI:        "git@github.com:..",
				PrivateKey: "private-key",
				WatchedPaths: []string{
					"watched/dir1",
					"watched/dir2",
				},
				IgnoredPaths: []string{
					"ignored/dir1/**",
					"README.md",
				},
				GitCryptKey:   "git-crypt-key",
				ManualTrigger: true,
			},
			TimerTrigger{
				Cron: "*/10 * * * *",
			},
			DockerTrigger{
				Image:    "ubuntu",
				Username: "userName",
				Password: "password",
			},
			PipelineTrigger{
				Pipeline: "a",
				Job:      "b",
				Status:   "c",
			},
		},
		Tasks: []Task{
			Run{
				Name:   "run task",
				Script: "script.sh --param",
				Docker: Docker{
					Image:    "golang:latest",
					Username: "user",
					Password: "pass",
				},
				Privileged: true,
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				SaveArtifacts: []string{
					"target/dist/artifact.zip",
					"README.md",
				},
				SaveArtifactsOnFailure: []string{
					"test_reports",
				},
			},
			DockerCompose{
				Name:         "docker compose task",
				ComposeFiles: []string{"../compose-file.yml"},
				Vars: Vars{
					"FOO":             "fOo",
					"BAR":             "1",
					"BAZ":             "true",
					"WRYY":            "2",
					"THIS_IS_STRANGE": "[a b]",
				},
				SaveArtifacts: []string{
					"target/dist/artifact.zip",
					"README.md",
				},
				SaveArtifactsOnFailure: []string{
					"test_reports",
				},
			},
			DockerPush{
				Name:                  "docker push task",
				Username:              "user",
				Password:              "pass",
				Image:                 "golang:latest",
				IgnoreVulnerabilities: true,
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				Notifications: Notifications{
					OnSuccess: []string{"asdf", "kehe"},
					OnFailure: []string{"kfds", "oasdf"},
				},
				Timeout: "1h",
				Tag:     "version",
			},
			DeployCF{
				Name:       "deploy cf task",
				API:        "cf.api",
				Space:      "cf.space",
				Org:        "cf.org",
				Rolling:    true,
				Username:   "cf.user",
				Password:   "cf.pass",
				TestDomain: "asdf.com",
				SSORoute:   "some.sso.route",
				Manifest:   "manifest.yml",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				DeployArtifact: "target/dist/artifact.zip",
				PreStart:       []string{"cf apps", "cf events"},
				PrePromote: []Task{
					Run{
						Script: "smoke-test.sh",
						Docker: Docker{
							Image: "golang",
						},
					},
					ConsumerIntegrationTest{
						Name:         "cdc-name",
						Consumer:     "cdc-consumer",
						ConsumerHost: "cdc-host",
						Script:       "cdc-script",
						UseCovenant:  false,
					}},
			},
			DeployKatee{
				Name:          "deploy katee task",
				ManualTrigger: false,
				Timeout:       "30s",
				Namespace:     "some-team",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				VelaManifest:    "blah",
				NotifyOnSuccess: false,
				Notifications: Notifications{
					OnSuccess: []string{"asdf", "kehe"},
					OnFailure: []string{"kfds", "oasdf"},
				},
				Tag: "latest",
			},
			DockerCompose{
				Name:    "docker compose task 2",
				Service: "asdf",
			},
			ConsumerIntegrationTest{
				Name:            "cdc-name",
				Consumer:        "cdc-consumer",
				ConsumerHost:    "cdc-host",
				GitCloneOptions: "--depth 100",
				Script:          "cdc-script",
				UseCovenant:     true,
			},
			DeployMLZip{
				Name:            "deploy ml zip",
				DeployZip:       "deploy-zip",
				AppName:         "app-name",
				AppVersion:      "app-version",
				Targets:         []string{"target1", "target2"},
				UseBuildVersion: true,
				Username:        "un",
				Password:        "pw",
			},
			DeployMLModules{
				MLModulesVersion: "ml-modules-version",
				AppName:          "app-name",
				AppVersion:       "app-version",
				Targets:          []string{"target1", "target2"},
				UseBuildVersion:  false,
				Password:         "p",
				BuildHistory:     10,
			},
			Parallel{
				Tasks: TaskList{
					Run{Name: "pr1"},
					Run{Name: "pr2"},
				},
			},
			Parallel{
				Tasks: TaskList{
					Sequence{
						Tasks: TaskList{
							Run{Name: "pr1"},
							Run{Name: "pr2"},
						},
					},
				},
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestInvalidYaml(t *testing.T) {
	yamls := []string{
		"team : { foo",
		"\t team: foo",
	}

	for _, yaml := range yamls {
		_, errs := Parse(yaml)
		assert.Equal(t, len(errs), 1, fmt.Sprintf("%q", yaml))
	}
}

func TestFailsWithUnknownFields(t *testing.T) {
	tests := []string{
		`
team: foo
tasks:
- type: run
  script: foo.sh
  docker:
    image: bash:latest
- type: docker-compose
  unknown_field: wibble`,
		`
team: foo
tasks:
- type: docker-compose
  unknown_field: wibble`,
		`
team: foo
triggers:
- type: git
  uri: git
  unknown_field: wobble

tasks:
- type: docker-compose`,
	}

	for i, yaml := range tests {
		_, errs := Parse(yaml)
		if assert.NotEmpty(t, errs, fmt.Sprintf("%v. %q", i, yaml)) {
			assert.Contains(t, fmt.Sprintf("%v", errs), "unknown_field")
		}
	}
}

func TestFailsWithDuplicateKeys(t *testing.T) {
	tests := []string{
		`
team: foo
platform: concourse
platform: actions
tasks:
- type: run
  script: foo.sh
  docker:
    image: bash:latest
- type: docker-compose
`,
		`
team: foo
tasks:
- type: docker-compose
  command: ./run
  command: ./run`,

		`
team: foo
tasks:
- type: parallel
  tasks:
    - type: run
      name: run1
    - type: run
      name: run2
  tasks:
     - type: run
       name: blah
`,
	}

	for i, yaml := range tests {
		_, errs := Parse(yaml)
		if assert.NotEmpty(t, errs, fmt.Sprintf("%v. %q", i, yaml)) {
			assert.Contains(t, fmt.Sprintf("%v", errs), "duplicated")
		}
	}
}

func TestGitTriggerShallowDefined(t *testing.T) {
	yaml := `
triggers:
- type: git
  shallow: false
`
	man, errs := Parse(yaml)
	assert.Empty(t, errs)
	assert.Equal(t, GitTrigger{ShallowDefined: true}, man.Triggers[0])
}

func TestDockerComposeIsSplitIntoArray(t *testing.T) {
	yaml := `
team: my team
pipeline: my pipeline
tasks:
- type: docker-compose
  compose_file: one two
`
	man, errs := Parse(yaml)
	assert.Empty(t, errs)

	expected := DockerCompose{ComposeFiles: []string{"one", "two"}}
	assert.Equal(t, expected, man.Tasks[0])
}
