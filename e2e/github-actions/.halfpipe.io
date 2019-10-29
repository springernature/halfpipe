team: engineering-enablement
pipeline: halfpipe-github-actions-run

feature_toggles:
- github-actions

triggers:
- type: git
  watched_paths:
  - e2e/github-actions
- type: timer
  cron: "* 10 * * *"

tasks:
- type: run
  name: test
  script: ./a
  docker:
    image: alpine:test
