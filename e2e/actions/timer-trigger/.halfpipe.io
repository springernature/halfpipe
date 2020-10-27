team: halfpipe-team
pipeline: timer-trigger

triggers:
- type: timer
  cron: '*/15 * * * *'


tasks:
- type: docker-push
  name: push to docker registry
  image: eu.gcr.io/halfpipe-io/someImage:someTag
