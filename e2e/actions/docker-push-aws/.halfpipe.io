team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: docker-push-aws
  name: Build and Push Docker Image
  image: example-app

- type: docker-push-aws
  name: Push with vars and secrets
  image: example-app-with-secrets
  vars:
    A: a
    B: b
  secrets:
    C: ((secret.c))
    D: d
