team: halfpipe-team
pipeline: halfpipe-e2e-docker-push

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-oci-push

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
  tag: "file:tagFile"
  ignore_vulnerabilities: true
