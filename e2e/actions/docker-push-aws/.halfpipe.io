team: halfpipe-team
pipeline: pipeline-name
platform: actions

tasks:
- type: docker-push-aws
  name: Build and Push Docker Image
  region: cn-northwest-1
  access_key_id: ((ee-private-ecr-user.aws_access_key_id))
  secret_access_key: ((ee-private-ecr-user.aws_secret_access_key))
  repository: example-app
