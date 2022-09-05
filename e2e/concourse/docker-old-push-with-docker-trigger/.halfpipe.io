team: halfpipe-team
pipeline: halfpipe-e2e-docker-push-with-docker-trigger

triggers:
- type: docker
  image: springernature/alpine:tag
- type: git
  watched_paths:
  - e2e/concourse/docker-old-push-with-docker-trigger

feature_toggles:
- docker-old-build

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

