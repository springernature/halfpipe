team: engineering-enablement
pipeline: halfpipe-e2e-docker-compose

feature_toggles:
- flatten-docker-compose

triggers:
- type: git
  watched_paths:
  - e2e/docker-compose

tasks:
- type: docker-compose
  name: test

- type: docker-compose
  name: two-services
  compose_file: docker-compose-2-services.yml
  service: customservice
  command: echo hello
  vars:
    F: foo
    B: bar
