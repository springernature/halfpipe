team: halfpipe-team
pipeline: halfpipe-e2e-docker-compose

feature_toggles:
- docker-decompose

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

- type: deploy-cf
  api: mp-api
  org: my-org
  space: my-space
  test_domain: test.com
  pre_promote:
    - type: docker-compose
      name: test2
