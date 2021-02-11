team: halfpipe-team
pipeline: docker-compose
platform: actions

tasks:
- type: docker-compose
  name: test
  save_artifacts:
  - foo
  - bar/baz
  save_artifacts_on_failure:
  - foo

- type: docker-compose
  name: custom
  compose_file: custom-docker-compose.yml
  service: customservice
  command: echo hello
  restore_artifacts: true
  vars:
    F: foo
    B: bar
