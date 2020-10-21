team: halfpipe-team
pipeline: halfpipe-e2e-update-pipeline

triggers:
- type: git
  watched_paths:
  - e2e/update-pipeline
- type: timer
  cron: '* * * * *'

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: Test
  script: a
  save_artifacts:
  - target/distribution
  - README.md
  save_artifacts_on_failure:
  - .halfpipe.io.yml
  docker:
    image: node:9.5.0-alpine

- type: deploy-cf
  name: deploy to cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  deploy_artifact: target/distribution/artifact.zip
  timeout: 5m

- type: parallel
  tasks:
  - type: deploy-cf
    name: deploy to staging
    api: ((cloudfoundry.api-snpaas))
    org: pe
    space: staging
    username: michiel
    password: very-secret
    vars:
        A: "0.1"
        B: "false"
    pre_promote:
    - type: run
      script: smoke-test.sh
      docker:
        image: eu.gcr.io/halfpipe-io/halfpipe-fly
      vars:
        A: "blah"
      restore_artifacts: true

    - type: consumer-integration-test
      name: c-name
      consumer: c-consumer
      consumer_host: c-host
      script: /var/c-script

    - type: docker-compose
      name: run pre promote step in docker-compose
      save_artifacts_on_failure:
        - path
  - type: deploy-cf
    name: deploy to qa
    api: ((cloudfoundry.api-snpaas))
    space: qa
    vars:
      A: "0.1"
      B: "false"
    pre_promote:
    - type: run
      name: save-artifact-in-pre-promote
      script: smoke-test.sh
      docker:
        image: eu.gcr.io/halfpipe-io/halfpipe-fly
      vars:
        A: "blah"

    - type: run
      name: restore artifact in pre promote
      script: smoke-test.sh
      docker:
        image: eu.gcr.io/halfpipe-io/halfpipe-fly
      vars:
        A: "blah"
      restore_artifacts: true

- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

- type: run
  script: ./notify.sh
  docker:
    image: busy
    username: michiel
    password: blah
  vars:
    A: a
    B: b

- type: docker-compose
  vars:
    A: a
  save_artifacts_on_failure:
    - docker-compose.yml

- type: consumer-integration-test
  name: another-c-name
  consumer: c-consumer
  consumer_host: c-host
  provider_host: p-host
  script: c-script
  docker_compose_service: potato
  vars:
    K: value
    K1: value1

- type: deploy-ml-zip
  deploy_zip: target/xquery.zip
  targets:
  - ml.dev.springer-sbm.com

- type: deploy-ml-modules
  name: Deploy ml-modules artifact
  ml_modules_version: "2.1425"
  app_name: my-app
  app_version: v1
  targets:
  - ml.dev.springer-sbm.com
  - ml.qa1.springer-sbm.com