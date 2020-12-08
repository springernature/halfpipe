team: halfpipe-team
pipeline: docker-compose

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
