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
  image: eu.gcr.io/halfpipe-io/someImage
  restore_artifacts: true

- type: docker-push
  name: Push custom
  image: dockerhubusername/someImage
  username: user
  password: ((foo.bar))
  restore_artifacts: true
  dockerfile_path: Dockerfile2
  timeout: 1h30m
  ignore_vulnerabilities: true
  vars:
    FOO: foo
    BAR: bar
    BLAH: ((very.secret))

- type: docker-push
  name: Push multiple platforms
  image: eu.gcr.io/halfpipe-io/someImage
  restore_artifacts: true
  platforms:
  - "linux/amd64"
  - "linux/arm64"

- type: docker-push
  name: Push multiple platforms and use cache
  image: eu.gcr.io/halfpipe-io/someImage
  restore_artifacts: true
  use_cache: true
  platforms:
  - "linux/amd64"
  - "linux/arm64"

- type: docker-push
  name: Push with secrets
  image: eu.gcr.io/halfpipe-io/someImage
  vars:
    A: a
    B: b
  secrets:
    C: ((secret.c))
    D: d
