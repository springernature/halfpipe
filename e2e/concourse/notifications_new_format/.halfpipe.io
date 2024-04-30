team: halfpipe-team
pipeline: halfpipe-e2e-notifications
notifications:
  failure:
  - slack: "#yo"

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
    success:
    - slack: asdf
      message: Wiiiie! \o/
    - slack: prws
      message: Wiiiie! \o/
    failure:
    - slack: kehe
      message: Nooooes >:c
    - slack: whoop
      message: Nooooes >:c
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

