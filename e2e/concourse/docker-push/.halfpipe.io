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
  image: springerplatformengineering/halfpipe_fly:thisIsMy_Tag
  vars:
    A: a
    B: b


- type: docker-push
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe_fly:thisIsMy_Tag2
  retries: 1
  ignore_vulnerabilities: true
  scan_timeout: 30

- type: docker-push
  image: springerplatformengineering/halfpipe_fly:thisIsMy_Tag2
  retries: 1
  ignore_vulnerabilities: true
  scan_timeout: 30
  use_cache: true
  platforms:
  - "linux/amd64"
  - "linux/arm64"

