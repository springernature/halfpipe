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
repo:
  uri: git@github.com:..
  private_key: private-key
  watched_paths:
  - watched/dir1
  - watched/dir2
  ignored_paths:
  - ignored/dir1/**
  - README.md
  git_crypt_key: git-crypt-key
slack_channel: "#ee-activity"
trigger_interval: 4h
artifact_config:
  bucket: myBucket
  json_key: ((some.jsonKey))
cron_trigger: "*/10 * * * *"
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
  vars:
    FOO: fOo
    BAR: "1"
  timeout: 1h
- type: deploy-cf
  name: deploy cf task
  api: cf.api
  space: cf.space
  org: cf.org
  username: cf.user
  password: cf.pass
  manifest: manifest.yml
  space: cf.space
  test_domain: asdf.com
  vars:
    FOO: fOo
    BAR: "1"
  deploy_artifact: target/dist/artifact.zip
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
- type: docker-compose
  name: docker compose task 2
  service: asdf
- type: consumer-integration-test
  name: cdc-name
  consumer: cdc-consumer
  consumer_host: cdc-host
  script: cdc-script
  parallel: true
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
- type: deploy-ml-modules
  app_name: app-name
  app_version: app-version
  ml_modules_version: ml-modules-version
  targets:
  - target1
  - target2
`)

	expected := Manifest{
		Team:     "my team",
		Pipeline: "my pipeline",
		Repo: Repo{
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
			GitCryptKey: "git-crypt-key",
		},
		ArtifactConfig: ArtifactConfig{
			Bucket:  "myBucket",
			JSONKey: "((some.jsonKey))",
		},
		SlackChannel:    "#ee-activity",
		TriggerInterval: "4h",
		CronTrigger:     "*/10 * * * *",
		FeatureToggles: FeatureToggles{
			"feature1",
			"feature2",
			"featureXYZ",
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
				Name:        "docker compose task",
				ComposeFile: "../compose-file.yml",
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
			DockerPush{
				Name:     "docker push task",
				Username: "user",
				Password: "pass",
				Image:    "golang:latest",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				Timeout: "1h",
			},
			DeployCF{
				Name:       "deploy cf task",
				API:        "cf.api",
				Space:      "cf.space",
				Org:        "cf.org",
				Username:   "cf.user",
				Password:   "cf.pass",
				TestDomain: "asdf.com",
				Manifest:   "manifest.yml",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				DeployArtifact: "target/dist/artifact.zip",
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
					}},
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
				Parallel:        true,
				Script:          "cdc-script",
			},
			DeployMLZip{
				Name:            "deploy ml zip",
				Parallel:        false,
				DeployZip:       "deploy-zip",
				AppName:         "app-name",
				AppVersion:      "app-version",
				Targets:         []string{"target1", "target2"},
				UseBuildVersion: true,
			},
			DeployMLModules{
				Parallel:         false,
				MLModulesVersion: "ml-modules-version",
				AppName:          "app-name",
				AppVersion:       "app-version",
				Targets:          []string{"target1", "target2"},
				UseBuildVersion:  false,
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
repo:
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
