team: halfpipe-team
pipeline: manual-git-trigger


triggers:
- type: git
  manual_trigger: true

tasks:
- type: docker-push
  name: push to docker registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
