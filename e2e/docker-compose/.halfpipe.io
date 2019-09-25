team: engineering-enablement
pipeline: halfpipe-e2e-docker-compose

triggers:
- type: git
  watched_paths:
  - e2e/docker-compose

tasks:
- type: docker-compose
  name: test

- type: docker-compose
  name: custom
  compose_file: custom-docker-compose.yml
  service: customservice
  command: echo hello
  vars:
    F: foo
    B: bar
