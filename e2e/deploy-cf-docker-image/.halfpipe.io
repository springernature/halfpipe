team: halfpipe-team
pipeline: halfpipe-e2e-deploy-cf-docker-image

feature_toggles:
- update-pipeline

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
  docker_tag: gitref
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
  docker_tag: version

- type: deploy-cf
  name: deploy to cf with pre promote - rolling deploy
  api: ((cloudfoundry.api-snpaas))
  space: dev
  rolling: true
  docker_tag: gitref
  pre_promote:
    - type: run
      name: pre promote step
      script: smoke-test.sh
      docker:
        image: eu.gcr.io/halfpipe-io/halfpipe-fly
      vars:
        A: "blah"
