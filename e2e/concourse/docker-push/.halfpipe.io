team: halfpipe-team
pipeline: halfpipe-e2e-docker-push

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-push

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

