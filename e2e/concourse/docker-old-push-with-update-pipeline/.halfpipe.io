team: halfpipe-team
pipeline: docker-push-with-update-pipeline

feature_toggles:
- update-pipeline
- docker-old-build

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-old-push-with-update-pipeline

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/image1
  vars:
    A: a
    B: b

- type: docker-push
  name: push to docker registry with git ref
  username: rob
  password: verysecret
  image: springerplatformengineering/image2
  vars:
    A: a
    B: b
  tag: gitref
