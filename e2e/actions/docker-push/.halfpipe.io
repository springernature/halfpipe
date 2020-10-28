team: halfpipe-team
pipeline: pipeline-name

tasks:
- type: docker-push
  name: Push to Docker Registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
  timeout: 1h30m
