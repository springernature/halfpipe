team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: run
  name: my run task
  docker:
    image: eu.gcr.io/halfpipe-io/golang:1.15
  script: \foo
  vars:
    FOO: foo
    BAR: bar
    SECRET1: ((something.cool))
    CUSTOM_PATH: ((/springernature/data/random/secret key))
    SECRET2: ((something.cooler))
    SHARED_SECRET: ((halfpipe-slack.token))
  timeout: 1h2m
- type: run
  docker:
    image: my.private.registry/repo/golang:1.15
    username: docker-user
    password: docker-password
  script: \bash -c "echo hello"
