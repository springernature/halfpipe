team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: docker-push-aws
  name: Build and Push Docker Image
  image: example-app
