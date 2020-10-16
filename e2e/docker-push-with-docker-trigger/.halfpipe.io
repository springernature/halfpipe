team: halfpipe-team
pipeline: halfpipe-e2e-docker-push-with-docker-trigger

triggers:
- type: docker
  image: springernature/alpine:tag
- type: git
  branch: 6.5.1
  watched_paths:
  - e2e/docker-push-with-docker-trigger

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

