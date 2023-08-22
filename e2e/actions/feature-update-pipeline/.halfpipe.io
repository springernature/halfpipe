team: halfpipe-team
pipeline: pipeline-name
platform: actions

feature_toggles:
- update-pipeline
- update-actions

tasks:
- type: run
  script: \echo hello
  docker:
    image: alpine
