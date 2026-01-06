team: halfpipe-team
pipeline: pipeline-ecr
platform: actions

tasks:
- type: docker-push
  name: Push to ECR
  image: 744877006609.dkr.ecr.cn-northwest-1.amazonaws.com.cn/ee-run/testrepo

