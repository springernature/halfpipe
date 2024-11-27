team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: docker-push
  name: Push custom
  image: eu.gcr.io/halfpipe-io/halfpipe-team/blah
  dockerfile_path: Dockerfile2
  build_path: context
  timeout: 1h30m
  ignore_vulnerabilities: true
  scan_timeout: 3
  vars:
    FOO: foo
    BAR: bar
    BLAH: ((very.secret))
  secrets:
    BLAH: ((very.secret))
  platforms:
  - linux/arm64
  - linux/amd64



