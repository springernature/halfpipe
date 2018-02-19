team: engineering-enablement

repo:
  uri: https://github.com/robwhitby/halfpipe-example-nodejs

tasks:
- name: run
  script: test.sh
  image: node:9.5.0-alpine

- name: deploy-cf
  api: ((cf-api.gcp))
  space: dev
  manifest: manifest.yml

- name: deploy-cf
  api: ((cf-api.dev))
  org: pe
  space: staging
