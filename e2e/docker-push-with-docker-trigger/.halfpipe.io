team: test
pipeline: test

triggers:
- type: docker
  image: alpine
- type: git
  watched_paths:
  - e2e/docker-push

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

