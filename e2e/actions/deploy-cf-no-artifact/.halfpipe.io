team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: deploy-cf
  api: ((cloudfoundry.api-snpaas))
  space: cf-space
