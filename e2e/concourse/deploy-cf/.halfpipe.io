team: halfpipe-team
pipeline: halfpipe-e2e-deploy-cf

triggers:
- type: git
  watched_paths:
  - e2e/deploy-cf

tasks:
- type: deploy-cf
  name: deploy to cf without any jazz
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  timeout: 5m
  cli_version: cf7

- type: deploy-cf
  name: deploy to cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  timeout: 5m
  pre_start:
  - cf apps
  - cf events myapp-CANDIDATE
  pre_promote:
  - type: run
    name: pre promote step
    script: smoke-test.sh
    docker:
      image: eu.gcr.io/halfpipe-io/halfpipe-fly
    vars:
      A: "blah"

