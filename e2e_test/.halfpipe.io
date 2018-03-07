team: engineering-enablement

repo:
  uri: https://github.com/robwhitby/halfpipe-example-nodejs

slack_channel: "#ee-re"

tasks:
- name: run
  script: test.sh
  docker:
    image: node:9.5.0-alpine

- name: deploy-cf
  api: dev-api
  space: dev
  manifest: manifest.yml
  username: michiel
  password: very-secret

- name: deploy-cf
  api: live-api
  org: pe
  space: staging
  username: michiel
  password: very-secret

- name: docker-push
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly

- name: run
  script: ./notify.sh
  docker:
    image: busy
    username: michiel
    password: blah