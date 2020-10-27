team: halfpipe-team
pipeline: git-trigger


triggers:
- type: git
  watched_paths:
  - e2e/actions/docker-push
  - e2e/actions

tasks:
- type: docker-push
  name: push to docker registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
