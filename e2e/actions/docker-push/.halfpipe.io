team: halfpipe-team
pipeline: docker-push

tasks:
- type: docker-push
  name: Push to Docker Registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
