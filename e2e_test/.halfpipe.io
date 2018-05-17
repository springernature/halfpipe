team: engineering-enablement
pipeline: e2e-test

repo:
  uri: https://github.com/robwhitby/halfpipe-example-nodejs

slack_channel: "#ee-re"

tasks:

- type: run
  name: Test
  script: test.sh
  save_artifacts:
  - target/distribution
  - README.md
  docker:
    image: node:9.5.0-alpine

- type: deploy-cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret
  test_domain: some.random.domain.com
  deploy_artifact: target/distribution/artifact.zip

- type: deploy-cf
  name: deploy to staging
  api: ((cloudfoundry.api-live))
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
  - type: consumer-integration-test
    name: c-name
    consumer: c-consumer
    host: c-host
    script: c-script

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

