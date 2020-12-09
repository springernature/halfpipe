team: halfpipe-team
pipeline: pipeline-name

tasks:
- type: run
  name: build
  docker:
    image: foo
  script: \build
  save_artifacts:
  - target/app.zip

- type: docker-push
  name: Push to Docker Registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
  restore_artifacts: true
  timeout: 1h30m

