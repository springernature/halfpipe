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
  tag: version

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
