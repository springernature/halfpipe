team: test
pipeline: halfpipe-e2e-notifications
slack_channel: "#yo"

triggers:
- type: git
  watched_paths:
  - e2e/notifications

tasks:
- type: run
  name: task1
  script: ./a
  docker:
    image: alpine
  notify_on_success: true

- type: run
  name: task2
  script: ./a
  docker:
    image: alpine

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
  notify_on_success: true
  pre_promote:
  - type: run
    script: ./a
    docker:
      image: eu.gcr.io/halfpipe-io/halfpipe-fly
    vars:
      A: "blah"

