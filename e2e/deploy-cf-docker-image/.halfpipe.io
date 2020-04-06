team: engineering-enablement
pipeline: halfpipe-e2e-deploy-cf-docker-image

triggers:
- type: git
  watched_paths:
  - e2e/deploy-cf-docker-image

tasks:
- type: deploy-cf
  name: deploy to cf simple
  api: ((cloudfoundry.api-snpaas))
  space: dev

- type: deploy-cf
  name: deploy to cf with pre promote
  api: ((cloudfoundry.api-snpaas))
  space: dev
  pre_promote:
  - type: run
    name: pre promote step
    script: smoke-test.sh
    docker:
      image: eu.gcr.io/halfpipe-io/halfpipe-fly
    vars:
      A: "blah"

- type: deploy-cf
  name: deploy to cf simple - rolling deploy
  api: ((cloudfoundry.api-snpaas))
  space: dev
  rolling: true

- type: deploy-cf
  name: deploy to cf with pre promote - rolling deploy
  api: ((cloudfoundry.api-snpaas))
  space: dev
  pre_promote:
    - type: run
      name: pre promote step
      script: smoke-test.sh
      docker:
        image: eu.gcr.io/halfpipe-io/halfpipe-fly
      vars:
        A: "blah"
