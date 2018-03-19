team: engineering-enablement

repo:
  uri: https://github.com/robwhitby/halfpipe-example-nodejs

slack_channel: "#ee-re"

tasks:
- type: run
  name: Test
  script: script-that-does-not-exist.sh
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
  deploy_artifact: target/distribution/artifact.zip

- type: deploy-cf
  name: deploy to staging
  api: live-api
  org: pe
  space: staging
  username: michiel
  password: very-secret
  vars:
      A: "0.1"
      B: "false"

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
