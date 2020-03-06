team: engineering-enablement
pipeline: halfpipe-e2e-deploy-cf-rolling

triggers:
- type: git
  watched_paths:
  - e2e/deploy-cf-rolling

tasks:
- type: deploy-cf
  rolling: true
  name: deploy to cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  timeout: 5m
  pre_promote:
  - type: run
    name: pre promote step
    script: smoke-test.sh
    docker:
      image: eu.gcr.io/halfpipe-io/halfpipe-fly
    vars:
      A: "blah"
