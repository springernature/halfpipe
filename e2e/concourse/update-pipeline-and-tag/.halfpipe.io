team: halfpipe-team
pipeline: pipeline-name

feature_toggles:
- update-pipeline-and-tag

tasks:
- type: run
  name: test
  script: ../update-pipeline/a
  docker:
    image: node:9.5.0-alpine
