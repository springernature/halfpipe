team: engineering-enablement
pipeline: halfpipe-e2e-deprecated-repo

repo:
  private_key: kehe
  watched_paths:
  - e2e/deprecated-repo

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  vars:
    A: a
    B: b

