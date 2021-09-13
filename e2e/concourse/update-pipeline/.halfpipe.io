team: halfpipe-team
pipeline: halfpipe-e2e-update-pipeline

triggers:
- type: git
  watched_paths:
  - e2e/concourse/update-pipeline
- type: timer
  cron: '*/30 * * * *'

feature_toggles:
- update-pipeline

tasks:
- type: run
  name: Test
  script: a
  docker:
    image: node:9.5.0-alpine

- type: docker-compose
