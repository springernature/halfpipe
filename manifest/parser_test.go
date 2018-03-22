package manifest

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestValidYaml_RequiredFields(t *testing.T) {
	man, errs := Parse(`
team: my team
pipeline: my pipeline
tasks:
- type: run
  script: run.sh
  docker:
    image: golang:latest
- type: docker-compose
- type: docker-push
  image: golang:latest
- type: deploy-cf
  api: ((cf.api))
  space: live
`)

	expected := Manifest{
		Team:     "my team",
		Pipeline: "my pipeline",
		Tasks: []Task{
			Run{
				Script: "run.sh",
				Docker: Docker{
					Image: "golang:latest",
				},
			},
			DockerCompose{},
			DockerPush{
				Image: "golang:latest",
			},
			DeployCF{
				API:   "((cf.api))",
				Space: "live",
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
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
- type: docker-compose
  name: docker compose task
  vars:
    FOO: fOo
    BAR: "1"
  save_artifacts:
  - target/dist/artifact.zip
  - README.md   
- type: docker-push
  name: docker push task
  username: user
  password: pass
  image: golang:latest
  vars:
    FOO: fOo
    BAR: "1"
- type: deploy-cf
  name: deploy cf task
  api: cf.api
  space: cf.space
  org: cf.org
  username: cf.user
  password: cf.pass
  manifest: manifest.yml
  space: cf.space
  vars:
    FOO: fOo
    BAR: "1"
  deploy_artifact: target/dist/artifact.zip
  pre_promote:
  - type: run
    script: smoke-test.sh
    docker:
      image: golang
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
		SlackChannel:    "#ee-activity",
		TriggerInterval: "4h",
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
			},
			DockerCompose{
				Name: "docker compose task",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				SaveArtifacts: []string{
					"target/dist/artifact.zip",
					"README.md",
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
			},
			DeployCF{
				Name:     "deploy cf task",
				API:      "cf.api",
				Space:    "cf.space",
				Org:      "cf.org",
				Username: "cf.user",
				Password: "cf.pass",
				Manifest: "manifest.yml",
				Vars: Vars{
					"FOO": "fOo",
					"BAR": "1",
				},
				DeployArtifact: "target/dist/artifact.zip",
				PrePromote: []Task{Run{
					Script: "smoke-test.sh",
					Docker: Docker{
						Image: "golang",
					},
				}},
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestInvalidYaml(t *testing.T) {
	yamls := []string{
		"team : { foo",
		" ",
		"\t team: foo",
	}

	for _, yaml := range yamls {
		_, errs := Parse(yaml)
		assert.Equal(t, len(errs), 1)
	}
}

func TestFailsWithUnknownFields(t *testing.T) {
	tests := []string{
		`
team: foo
tasks:
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

	for _, test := range tests {
		_, errs := Parse(test)
		if assert.NotEmpty(t, errs) {
			assert.Contains(t, fmt.Sprintf("%v", errs), "unknown_field")
		}
	}
}
