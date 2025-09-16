team: halfpipe-team
pipeline: pipeline-name
platform: actions

feature_toggles:
- ghas

tasks:
- type: docker-push
  name: Push default
  image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
