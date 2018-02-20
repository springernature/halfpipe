team: engineering-enablement

repo:
  uri: https://github.com/robwhitby/halfpipe-example-nodejs
  git_crypt_key: can't do secrets in e2e test :(

tasks:
- name: run
  script: test.sh
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
  repo: springerplatformengineering/halfpipe-fly

- name: run
  script: ./notify.sh
  image: busybox