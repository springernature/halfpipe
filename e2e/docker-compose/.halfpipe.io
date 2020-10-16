team: halfpipe-team
pipeline: halfpipe-e2e-docker-compose

triggers:
- type: git
  branch: 6.5.1
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
