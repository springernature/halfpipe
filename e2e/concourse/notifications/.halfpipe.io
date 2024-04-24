team: halfpipe-team
pipeline: halfpipe-e2e-notifications
slack_channel: "#yo"
teams_webhook: "https://someHorribleLongURL"

triggers:
- type: git
  watched_paths:
  - e2e/concourse/notifications

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
  notifications:
    on_success_message: Wiiiie! \o/
    on_success:
    - asdf
    - prws
    on_failure_message: Nooooes >:c
    on_failure:
    - kehe
    - whoop
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
  notify_on_success: true
  pre_promote:
  - type: run
    script: ./a
    docker:
      image: eu.gcr.io/halfpipe-io/halfpipe-fly
    vars:
      A: "blah"

