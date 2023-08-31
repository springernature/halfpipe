team: halfpipe-team
pipeline: pipeline-name
platform: actions

feature_toggles:
- update-pipeline

tasks:
- type: run
  script: \echo hello
  docker:
    image: alpine
