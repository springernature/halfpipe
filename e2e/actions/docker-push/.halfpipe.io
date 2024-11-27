team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: docker
  image: eu.gcr.io/halfpipe-io/baseImage

tasks:
- type: run
  name: build
  docker:
    image: foo
  script: \build
  save_artifacts:
  - target/app.zip

- type: docker-push
  name: Push default
  image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
  restore_artifacts: true

- type: docker-push
  name: Push multiple platforms
  image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
  restore_artifacts: true
  platforms:
  - "linux/amd64"
  - "linux/arm64"

- type: docker-push
  name: Push with secrets
  image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
  vars:
    A: a
    B: b
  secrets:
    C: ((secret.c))
    D: d
