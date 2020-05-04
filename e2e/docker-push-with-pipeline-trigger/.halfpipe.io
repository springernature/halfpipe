team: halfpipe-team
pipeline: halfpipe-e2e-docker-push-with-pipeline-trigger

triggers:
- type: git
  watched_paths:
  - e2e/docker-push-with-pipeline-trigger
- type: pipeline
  pipeline: halfpipe-example-docker-push
  job: docker-push

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

