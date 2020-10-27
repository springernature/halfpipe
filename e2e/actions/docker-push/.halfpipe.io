team: halfpipe-team
pipeline: docker-push-test

tasks:
  - type: docker-push
    name: push to docker registry
    image: eu.gcr.io/halfpipe-io/someImage:someTag
