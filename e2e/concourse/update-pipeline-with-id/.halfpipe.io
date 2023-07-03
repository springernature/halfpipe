team: halfpipe-team
pipeline: halfpipe-e2e-update-pipeline-with-id
pipeline_id: my-pipeline-id

feature_toggles:
- update-pipeline

tasks:
- type: run
  script: \echo hello
  docker:
    image: alpine
