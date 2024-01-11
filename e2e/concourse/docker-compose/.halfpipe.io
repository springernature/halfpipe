team: halfpipe-team
pipeline: halfpipe-e2e-docker-compose

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-compose

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

- type: docker-compose
  name: multiple-compose-files
  command: echo hello
  compose_file: docker-compose.yml custom-docker-compose.yml
