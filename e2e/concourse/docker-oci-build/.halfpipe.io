team: halfpipe-team
pipeline: halfpipe-e2e-docker-oci-build

feature_toggles:
- docker-oci-build

triggers:
- type: git
  watched_paths:
  - e2e/concourse/docker-oci-build

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe_fly:thisIsMy_Tag
