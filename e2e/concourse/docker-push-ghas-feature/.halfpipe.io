team: halfpipe-team
pipeline: pipeline-name

feature_toggles:
- ghas

tasks:
- type: docker-push
  image: eu.gcr.io/halfpipe-io/halfpipe-team/someImage
