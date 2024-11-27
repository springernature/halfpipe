team: halfpipe-team
pipeline: halfpipe-e2e-artifacts

triggers:
- type: git
  watched_paths:
  - .

tasks:
- type: run
  name: create-artifact
  script: ./a
  docker:
    image: alpine
  save_artifacts:
  - someFile
  - ../parentDir/someFile2
  save_artifacts_on_failure:
  - .halfpipe.io.yml
  - ../../.halfpipe.io.yml

- type: run
  name: read-artifact
  script: ./a
  docker:
    image: alpine
  restore_artifacts: true

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
    script: ./a
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
  restore_artifacts: true

- type: docker-compose
  vars:
    A: a
  save_artifacts_on_failure:
    - docker-compose.yml
  restore_artifacts: true

- type: deploy-ml-zip
  deploy_zip: target/xquery.zip
  targets:
  - ml.dev.springer-sbm.com


- type: consumer-integration-test
  name: c-name-covenant
  consumer: c-consumer
  consumer_host: c-host
  script: c-script
  save_artifacts:
  - .
  save_artifacts_on_failure:
  - docker-compose.yml
  - tests

